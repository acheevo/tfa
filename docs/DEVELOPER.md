# Developer Guide

This guide helps developers understand how to extend, customize, and contribute to the Fullstack Template. It covers patterns, conventions, and best practices used throughout the codebase.

## Table of Contents

- [Getting Started](#getting-started)
- [Code Organization](#code-organization)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Testing Guidelines](#testing-guidelines)
- [Adding New Features](#adding-new-features)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Getting Started

### Development Setup

1. **Clone and install dependencies**:
   ```bash
   git clone <repository-url>
   cd fullstack-template
   make install
   ```

2. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Start development servers**:
   ```bash
   # Terminal 1: Frontend dev server (hot reload)
   make frontend-dev
   
   # Terminal 2: Backend API server
   make dev
   ```

4. **Run tests**:
   ```bash
   # Backend tests
   make test
   
   # Frontend tests
   cd frontend && npm test
   ```

### Development Workflow

1. **Create feature branch**: `git checkout -b feature/feature-name`
2. **Make changes**: Follow coding standards and patterns
3. **Add tests**: Ensure good test coverage
4. **Run quality checks**: `make lint && make frontend-lint`
5. **Test thoroughly**: `make test && cd frontend && npm test`
6. **Commit changes**: Use conventional commit messages
7. **Create pull request**: Include description and testing notes

---

## Code Organization

### Project Structure Philosophy

The codebase follows these organizational principles:

1. **Domain-Driven Design**: Code organized by business domains
2. **Clean Architecture**: Clear separation of concerns
3. **Convention over Configuration**: Consistent patterns throughout
4. **Modular Design**: Easy to add/remove features

### Backend Organization

```
internal/
├── auth/                   # Authentication domain
│   ├── domain/            # Business entities and rules
│   ├── repository/        # Data access implementations
│   ├── service/           # Business logic
│   └── transport/         # HTTP handlers
├── user/                  # User management domain
├── admin/                 # Admin operations domain
├── middleware/            # Cross-cutting HTTP middleware
│   ├── auth.go           # Authentication middleware
│   ├── rbac.go           # Authorization middleware
│   ├── cors.go           # CORS handling
│   ├── logger.go         # Request logging
│   └── rate_limit.go     # Rate limiting
└── shared/               # Shared utilities
    ├── config/           # Configuration management
    ├── database/         # Database setup and migrations
    ├── email/            # Email service
    ├── logger/           # Structured logging
    └── monitoring/       # Health checks and metrics
```

### Frontend Organization

```
src/
├── components/           # React components
│   ├── ui/              # Generic, reusable components
│   ├── auth/            # Authentication-specific components
│   ├── admin/           # Admin-specific components
│   └── layout/          # Layout components
├── contexts/            # React contexts for global state
├── hooks/               # Custom React hooks
├── lib/                 # Utilities and services
├── pages/               # Page-level components
├── types/               # TypeScript type definitions
└── test/                # Test utilities and mocks
```

---

## Backend Development

### Adding a New Domain Module

Follow this pattern when adding new business domains:

1. **Create domain structure**:
   ```bash
   mkdir -p internal/product/{domain,repository,service,transport}
   ```

2. **Define domain entities** (`internal/product/domain/types.go`):
   ```go
   package domain

   import (
       "time"
       "gorm.io/gorm"
   )

   type Product struct {
       ID          uint           `json:"id" gorm:"primarykey"`
       Name        string         `json:"name" gorm:"not null"`
       Description string         `json:"description"`
       Price       float64        `json:"price" gorm:"not null"`
       UserID      uint           `json:"user_id" gorm:"not null"`
       CreatedAt   time.Time      `json:"created_at"`
       UpdatedAt   time.Time      `json:"updated_at"`
       DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
   }

   type CreateProductRequest struct {
       Name        string  `json:"name" binding:"required,min=1,max=100"`
       Description string  `json:"description" binding:"max=500"`
       Price       float64 `json:"price" binding:"required,min=0"`
   }

   type UpdateProductRequest struct {
       Name        string  `json:"name" binding:"omitempty,min=1,max=100"`
       Description string  `json:"description" binding:"omitempty,max=500"`
       Price       float64 `json:"price" binding:"omitempty,min=0"`
   }
   ```

3. **Create repository interface** (`internal/product/repository/product.go`):
   ```go
   package repository

   import (
       "gorm.io/gorm"
       "github.com/your-org/fullstack-template/internal/product/domain"
   )

   type ProductRepository interface {
       Create(product *domain.Product) error
       GetByID(id uint) (*domain.Product, error)
       GetByUserID(userID uint) ([]*domain.Product, error)
       Update(product *domain.Product) error
       Delete(id uint) error
   }

   type productRepository struct {
       db *gorm.DB
   }

   func NewProductRepository(db *gorm.DB) ProductRepository {
       return &productRepository{db: db}
   }

   func (r *productRepository) Create(product *domain.Product) error {
       return r.db.Create(product).Error
   }

   func (r *productRepository) GetByID(id uint) (*domain.Product, error) {
       var product domain.Product
       err := r.db.First(&product, id).Error
       if err != nil {
           if err == gorm.ErrRecordNotFound {
               return nil, domain.ErrProductNotFound
           }
           return nil, err
       }
       return &product, nil
   }
   ```

4. **Implement service layer** (`internal/product/service/product.go`):
   ```go
   package service

   import (
       "log/slog"
       "github.com/your-org/fullstack-template/internal/product/domain"
       "github.com/your-org/fullstack-template/internal/product/repository"
       "github.com/your-org/fullstack-template/internal/shared/config"
   )

   type ProductService struct {
       config     *config.Config
       logger     *slog.Logger
       productRepo repository.ProductRepository
   }

   func NewProductService(
       config *config.Config,
       logger *slog.Logger,
       productRepo repository.ProductRepository,
   ) *ProductService {
       return &ProductService{
           config:     config,
           logger:     logger,
           productRepo: productRepo,
       }
   }

   func (s *ProductService) CreateProduct(userID uint, req *domain.CreateProductRequest) (*domain.Product, error) {
       product := &domain.Product{
           Name:        req.Name,
           Description: req.Description,
           Price:       req.Price,
           UserID:      userID,
       }

       if err := s.productRepo.Create(product); err != nil {
           s.logger.Error("failed to create product", "error", err)
           return nil, err
       }

       s.logger.Info("product created", "product_id", product.ID, "user_id", userID)
       return product, nil
   }
   ```

5. **Add HTTP handlers** (`internal/product/transport/http.go`):
   ```go
   package transport

   import (
       "net/http"
       "strconv"
       "log/slog"

       "github.com/gin-gonic/gin"
       "github.com/your-org/fullstack-template/internal/product/domain"
       "github.com/your-org/fullstack-template/internal/product/service"
       "github.com/your-org/fullstack-template/internal/middleware"
   )

   type ProductHandler struct {
       productService *service.ProductService
       logger         *slog.Logger
   }

   func NewProductHandler(productService *service.ProductService, logger *slog.Logger) *ProductHandler {
       return &ProductHandler{
           productService: productService,
           logger:         logger,
       }
   }

   func (h *ProductHandler) CreateProduct(c *gin.Context) {
       var req domain.CreateProductRequest
       if err := c.ShouldBindJSON(&req); err != nil {
           c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
           return
       }

       userID := middleware.GetUserIDFromContext(c)
       product, err := h.productService.CreateProduct(userID, &req)
       if err != nil {
           c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
           return
       }

       c.JSON(http.StatusCreated, product)
   }

   func RegisterRoutes(r *gin.RouterGroup, productService *service.ProductService, logger *slog.Logger) {
       handler := NewProductHandler(productService, logger)
       
       products := r.Group("/products")
       products.Use(middleware.RequireAuth())
       {
           products.POST("", handler.CreateProduct)
           products.GET("", handler.GetUserProducts)
           products.GET("/:id", handler.GetProduct)
           products.PUT("/:id", handler.UpdateProduct)
           products.DELETE("/:id", handler.DeleteProduct)
       }
   }
   ```

6. **Register in main server** (`internal/http/server.go`):
   ```go
   // Add to imports
   producttransport "github.com/your-org/fullstack-template/internal/product/transport"
   productservice "github.com/your-org/fullstack-template/internal/product/service"
   productrepo "github.com/your-org/fullstack-template/internal/product/repository"

   // Add to server setup
   func NewServer(cfg *config.Config, logger *slog.Logger, db *gorm.DB) *Server {
       // ... existing code ...

       // Product module
       productRepo := productrepo.NewProductRepository(db)
       productService := productservice.NewProductService(cfg, logger, productRepo)

       // Register routes
       api := r.Group("/api")
       producttransport.RegisterRoutes(api, productService, logger)
   }
   ```

### Database Migrations

Add new entities to auto-migration:

```go
// internal/shared/database/database.go
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &authdomain.User{},
        &authdomain.RefreshToken{},
        &authdomain.PasswordReset{},
        &authdomain.AuditLog{},
        &productdomain.Product{}, // Add new entity
    )
}
```

### Error Handling Patterns

Follow consistent error handling:

```go
// Domain errors (internal/product/domain/errors.go)
var (
    ErrProductNotFound     = errors.New("product not found")
    ErrProductUnauthorized = errors.New("unauthorized to access product")
    ErrInvalidPrice        = errors.New("price must be positive")
)

// Service layer error handling
func (s *ProductService) GetProduct(userID, productID uint) (*domain.Product, error) {
    product, err := s.productRepo.GetByID(productID)
    if err != nil {
        if err == domain.ErrProductNotFound {
            return nil, err
        }
        s.logger.Error("failed to get product", "product_id", productID, "error", err)
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    // Check ownership
    if product.UserID != userID {
        return nil, domain.ErrProductUnauthorized
    }

    return product, nil
}

// HTTP layer error handling
func (h *ProductHandler) GetProduct(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := middleware.GetUserIDFromContext(c)

    product, err := h.productService.GetProduct(userID, uint(id))
    if err != nil {
        switch err {
        case domain.ErrProductNotFound:
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        case domain.ErrProductUnauthorized:
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
        default:
            h.logger.Error("get product failed", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }

    c.JSON(http.StatusOK, product)
}
```

### Testing Backend Code

Create comprehensive tests for each layer:

```go
// Test service layer (internal/product/service/product_test.go)
func TestProductService_CreateProduct(t *testing.T) {
    tests := []struct {
        name    string
        userID  uint
        request *domain.CreateProductRequest
        want    *domain.Product
        wantErr bool
    }{
        {
            name:   "successful creation",
            userID: 1,
            request: &domain.CreateProductRequest{
                Name:        "Test Product",
                Description: "Test Description",
                Price:       99.99,
            },
            want: &domain.Product{
                Name:        "Test Product",
                Description: "Test Description", 
                Price:       99.99,
                UserID:      1,
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockRepo := &MockProductRepository{}
            service := NewProductService(nil, nil, mockRepo)

            // Execute
            got, err := service.CreateProduct(tt.userID, tt.request)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.Name, got.Name)
            }
        })
    }
}
```

---

## Frontend Development

### Adding New Components

Follow these patterns when creating React components:

1. **UI Components** (generic, reusable):
   ```typescript
   // src/components/ui/Card.tsx
   import React from 'react';
   import { cn } from '../../utils/cn';

   interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
     children: React.ReactNode;
     variant?: 'default' | 'outlined' | 'elevated';
   }

   export const Card: React.FC<CardProps> = ({ 
     children, 
     variant = 'default', 
     className,
     ...props 
   }) => {
     return (
       <div
         className={cn(
           'rounded-lg p-4',
           {
             'bg-white shadow-sm': variant === 'default',
             'border border-gray-200': variant === 'outlined',
             'bg-white shadow-lg': variant === 'elevated',
           },
           className
         )}
         {...props}
       >
         {children}
       </div>
     );
   };
   ```

2. **Domain Components** (business-specific):
   ```typescript
   // src/components/product/ProductCard.tsx
   import React from 'react';
   import { Card } from '../ui/Card';
   import { Button } from '../ui/Button';
   import { Product } from '../../types/product';

   interface ProductCardProps {
     product: Product;
     onEdit?: (product: Product) => void;
     onDelete?: (product: Product) => void;
     readonly?: boolean;
   }

   export const ProductCard: React.FC<ProductCardProps> = ({
     product,
     onEdit,
     onDelete,
     readonly = false,
   }) => {
     return (
       <Card className="space-y-4">
         <div>
           <h3 className="text-lg font-semibold">{product.name}</h3>
           <p className="text-gray-600">{product.description}</p>
           <p className="text-xl font-bold">${product.price}</p>
         </div>

         {!readonly && (
           <div className="flex gap-2">
             {onEdit && (
               <Button 
                 variant="outline" 
                 onClick={() => onEdit(product)}
               >
                 Edit
               </Button>
             )}
             {onDelete && (
               <Button 
                 variant="destructive" 
                 onClick={() => onDelete(product)}
               >
                 Delete
               </Button>
             )}
           </div>
         )}
       </Card>
     );
   };
   ```

3. **Form Components**:
   ```typescript
   // src/components/product/ProductForm.tsx
   import React, { useState } from 'react';
   import { Button } from '../ui/Button';
   import { Input } from '../ui/Input';
   import { CreateProductRequest, Product } from '../../types/product';

   interface ProductFormProps {
     initialData?: Product;
     onSubmit: (data: CreateProductRequest) => Promise<void>;
     onCancel?: () => void;
     isLoading?: boolean;
   }

   export const ProductForm: React.FC<ProductFormProps> = ({
     initialData,
     onSubmit,
     onCancel,
     isLoading = false,
   }) => {
     const [formData, setFormData] = useState<CreateProductRequest>({
       name: initialData?.name || '',
       description: initialData?.description || '',
       price: initialData?.price || 0,
     });

     const [errors, setErrors] = useState<Record<string, string>>({});

     const handleSubmit = async (e: React.FormEvent) => {
       e.preventDefault();
       
       // Validation
       const newErrors: Record<string, string> = {};
       if (!formData.name.trim()) {
         newErrors.name = 'Name is required';
       }
       if (formData.price <= 0) {
         newErrors.price = 'Price must be positive';
       }

       if (Object.keys(newErrors).length > 0) {
         setErrors(newErrors);
         return;
       }

       try {
         await onSubmit(formData);
         setErrors({});
       } catch (error) {
         // Handle submission errors
         console.error('Form submission error:', error);
       }
     };

     return (
       <form onSubmit={handleSubmit} className="space-y-4">
         <Input
           label="Product Name"
           value={formData.name}
           onChange={(e) => setFormData({ ...formData, name: e.target.value })}
           error={errors.name}
           disabled={isLoading}
         />

         <Input
           label="Description"
           value={formData.description}
           onChange={(e) => setFormData({ ...formData, description: e.target.value })}
           error={errors.description}
           disabled={isLoading}
         />

         <Input
           label="Price"
           type="number"
           step="0.01"
           min="0"
           value={formData.price}
           onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) })}
           error={errors.price}
           disabled={isLoading}
         />

         <div className="flex gap-2">
           <Button type="submit" loading={isLoading}>
             {initialData ? 'Update' : 'Create'} Product
           </Button>
           {onCancel && (
             <Button type="button" variant="outline" onClick={onCancel}>
               Cancel
             </Button>
           )}
         </div>
       </form>
     );
   };
   ```

### Custom Hooks

Create reusable logic with custom hooks:

```typescript
// src/hooks/useProducts.ts
import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Product, CreateProductRequest } from '../types/product';

export const useProducts = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProducts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getProducts();
      setProducts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch products');
    } finally {
      setLoading(false);
    }
  };

  const createProduct = async (data: CreateProductRequest) => {
    try {
      const newProduct = await api.createProduct(data);
      setProducts(prev => [...prev, newProduct]);
      return newProduct;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create product');
      throw err;
    }
  };

  const updateProduct = async (id: number, data: Partial<Product>) => {
    try {
      const updatedProduct = await api.updateProduct(id, data);
      setProducts(prev => 
        prev.map(p => p.id === id ? updatedProduct : p)
      );
      return updatedProduct;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update product');
      throw err;
    }
  };

  const deleteProduct = async (id: number) => {
    try {
      await api.deleteProduct(id);
      setProducts(prev => prev.filter(p => p.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete product');
      throw err;
    }
  };

  useEffect(() => {
    fetchProducts();
  }, []);

  return {
    products,
    loading,
    error,
    createProduct,
    updateProduct,
    deleteProduct,
    refetch: fetchProducts,
  };
};
```

### State Management Patterns

Use Context for global state:

```typescript
// src/contexts/ProductContext.tsx
import React, { createContext, useContext, ReactNode } from 'react';
import { useProducts } from '../hooks/useProducts';
import { Product, CreateProductRequest } from '../types/product';

interface ProductContextType {
  products: Product[];
  loading: boolean;
  error: string | null;
  createProduct: (data: CreateProductRequest) => Promise<Product>;
  updateProduct: (id: number, data: Partial<Product>) => Promise<Product>;
  deleteProduct: (id: number) => Promise<void>;
  refetch: () => Promise<void>;
}

const ProductContext = createContext<ProductContextType | undefined>(undefined);

export const ProductProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const productState = useProducts();

  return (
    <ProductContext.Provider value={productState}>
      {children}
    </ProductContext.Provider>
  );
};

export const useProductContext = (): ProductContextType => {
  const context = useContext(ProductContext);
  if (context === undefined) {
    throw new Error('useProductContext must be used within a ProductProvider');
  }
  return context;
};
```

### Testing Frontend Components

Create comprehensive component tests:

```typescript
// src/components/product/ProductCard.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProductCard } from './ProductCard';
import { Product } from '../../types/product';

const mockProduct: Product = {
  id: 1,
  name: 'Test Product',
  description: 'Test Description',
  price: 99.99,
  user_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('ProductCard', () => {
  it('renders product information', () => {
    render(<ProductCard product={mockProduct} />);
    
    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('Test Description')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
  });

  it('calls onEdit when edit button is clicked', () => {
    const onEdit = vi.fn();
    render(<ProductCard product={mockProduct} onEdit={onEdit} />);
    
    fireEvent.click(screen.getByText('Edit'));
    expect(onEdit).toHaveBeenCalledWith(mockProduct);
  });

  it('calls onDelete when delete button is clicked', () => {
    const onDelete = vi.fn();
    render(<ProductCard product={mockProduct} onDelete={onDelete} />);
    
    fireEvent.click(screen.getByText('Delete'));
    expect(onDelete).toHaveBeenCalledWith(mockProduct);
  });

  it('hides action buttons in readonly mode', () => {
    render(
      <ProductCard 
        product={mockProduct} 
        onEdit={vi.fn()} 
        onDelete={vi.fn()} 
        readonly 
      />
    );
    
    expect(screen.queryByText('Edit')).not.toBeInTheDocument();
    expect(screen.queryByText('Delete')).not.toBeInTheDocument();
  });
});
```

---

## Testing Guidelines

### Backend Testing Strategy

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test API endpoints with database
3. **Repository Tests**: Test data access layer
4. **Service Tests**: Test business logic

### Frontend Testing Strategy

1. **Component Tests**: Test component behavior and rendering
2. **Hook Tests**: Test custom hook logic
3. **Integration Tests**: Test component interactions
4. **API Tests**: Test API client functionality

### Test Organization

```
internal/
├── auth/
│   ├── service/
│   │   ├── auth.go
│   │   └── auth_test.go          # Unit tests
│   └── repository/
│       ├── user.go
│       └── user_test.go          # Repository tests
└── test/
    └── integration/              # Integration tests
        ├── auth_test.go
        ├── user_test.go
        └── setup_test.go

frontend/src/
├── components/
│   └── ui/
│       ├── Button.tsx
│       └── Button.test.tsx       # Component tests
├── hooks/
│   ├── useAuth.ts
│   └── useAuth.test.ts           # Hook tests
└── test/
    ├── setup.ts                  # Test setup
    ├── mocks/                    # Mock data and services
    └── utils/                    # Test utilities
```

### Writing Good Tests

1. **Follow AAA Pattern**: Arrange, Act, Assert
2. **Test Behavior, Not Implementation**: Focus on what, not how
3. **Use Descriptive Names**: Test names should explain the scenario
4. **Keep Tests Independent**: Each test should be isolated
5. **Mock External Dependencies**: Use mocks for databases, APIs

Example test patterns:

```go
// Good: Tests behavior with clear scenario
func TestAuthService_Login_WithValidCredentials_ReturnsTokens(t *testing.T) {
    // Arrange
    user := createTestUser(t)
    mockRepo := &MockUserRepository{}
    service := NewAuthService(mockRepo)

    // Act
    response, err := service.Login(&LoginRequest{
        Email:    user.Email,
        Password: "password123",
    })

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, response.AccessToken)
    assert.NotEmpty(t, response.RefreshToken)
}

// Bad: Vague test name, testing implementation
func TestLogin(t *testing.T) {
    service := NewAuthService()
    service.userRepo.GetByEmail("test@example.com")
    // ...
}
```

---

## Adding New Features

### Feature Development Checklist

When adding a new feature, follow this checklist:

- [ ] **Design**: Plan the feature architecture
- [ ] **Backend Implementation**:
  - [ ] Domain entities and types
  - [ ] Repository interfaces and implementations
  - [ ] Service layer with business logic
  - [ ] HTTP handlers and routing
  - [ ] Database migrations
  - [ ] Error handling
  - [ ] Input validation
  - [ ] Unit tests
  - [ ] Integration tests
- [ ] **Frontend Implementation**:
  - [ ] TypeScript types
  - [ ] API client methods
  - [ ] React components
  - [ ] Custom hooks
  - [ ] State management
  - [ ] Form validation
  - [ ] Component tests
  - [ ] Integration tests
- [ ] **Documentation**:
  - [ ] API documentation
  - [ ] Component documentation
  - [ ] Usage examples
- [ ] **Quality Assurance**:
  - [ ] Code review
  - [ ] Manual testing
  - [ ] Performance testing
  - [ ] Security review

### Example: Adding a Notification System

1. **Backend Implementation**:
   ```go
   // Domain
   type Notification struct {
       ID       uint      `json:"id"`
       UserID   uint      `json:"user_id"`
       Type     string    `json:"type"`
       Title    string    `json:"title"`
       Message  string    `json:"message"`
       Read     bool      `json:"read"`
       Priority string    `json:"priority"`
   }

   // Service
   func (s *NotificationService) CreateNotification(userID uint, req *CreateNotificationRequest) error {
       notification := &Notification{
           UserID:   userID,
           Type:     req.Type,
           Title:    req.Title,
           Message:  req.Message,
           Priority: req.Priority,
       }
       return s.repo.Create(notification)
   }

   // HTTP Handler
   func (h *NotificationHandler) GetNotifications(c *gin.Context) {
       userID := middleware.GetUserIDFromContext(c)
       notifications, err := h.service.GetUserNotifications(userID)
       if err != nil {
           c.JSON(500, gin.H{"error": "Failed to get notifications"})
           return
       }
       c.JSON(200, notifications)
   }
   ```

2. **Frontend Implementation**:
   ```typescript
   // Hook
   export const useNotifications = () => {
     const [notifications, setNotifications] = useState<Notification[]>([]);
     
     const fetchNotifications = async () => {
       const data = await api.getNotifications();
       setNotifications(data);
     };

     const markAsRead = async (id: number) => {
       await api.markNotificationAsRead(id);
       setNotifications(prev => 
         prev.map(n => n.id === id ? { ...n, read: true } : n)
       );
     };

     return { notifications, fetchNotifications, markAsRead };
   };

   // Component
   export const NotificationList: React.FC = () => {
     const { notifications, markAsRead } = useNotifications();

     return (
       <div className="space-y-2">
         {notifications.map(notification => (
           <NotificationItem
             key={notification.id}
             notification={notification}
             onMarkAsRead={markAsRead}
           />
         ))}
       </div>
     );
   };
   ```

---

## Best Practices

### Code Quality

1. **Follow Language Conventions**:
   - Go: Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - TypeScript: Follow [TypeScript Handbook](https://www.typescriptlang.org/docs/)

2. **Use Consistent Naming**:
   - Go: `CamelCase` for public, `camelCase` for private
   - TypeScript: `PascalCase` for components, `camelCase` for functions

3. **Write Self-Documenting Code**:
   ```go
   // Good
   func calculateMonthlyInterest(principal float64, annualRate float64) float64 {
       return principal * (annualRate / 12)
   }

   // Bad
   func calc(p, r float64) float64 {
       return p * (r / 12)
   }
   ```

4. **Handle Errors Properly**:
   ```go
   // Good
   user, err := userService.GetUser(id)
   if err != nil {
       logger.Error("failed to get user", "user_id", id, "error", err)
       return fmt.Errorf("get user failed: %w", err)
   }

   // Bad
   user, _ := userService.GetUser(id)
   ```

### Security Best Practices

1. **Input Validation**:
   - Always validate and sanitize user input
   - Use binding validation tags in Go
   - Validate on both frontend and backend

2. **Authentication**:
   - Use secure JWT tokens
   - Implement proper token refresh
   - Set appropriate token expiration

3. **Authorization**:
   - Check permissions at every endpoint
   - Use role-based access control
   - Implement principle of least privilege

4. **Data Protection**:
   - Hash passwords with bcrypt
   - Use HTTPS in production
   - Sanitize data for XSS prevention

### Performance Best Practices

1. **Database Optimization**:
   - Add appropriate indexes
   - Use pagination for large datasets
   - Optimize queries with EXPLAIN

2. **Frontend Optimization**:
   - Use React.memo for expensive components
   - Implement code splitting
   - Optimize bundle size

3. **Caching**:
   - Cache static data appropriately
   - Use HTTP caching headers
   - Implement application-level caching

---

## Troubleshooting

### Common Backend Issues

1. **Database Connection Errors**:
   ```bash
   # Check PostgreSQL status
   pg_isready -h localhost -p 5432
   
   # Verify connection string
   echo $DATABASE_URL
   
   # Test manual connection
   psql -h localhost -U postgres -d fullstack_template
   ```

2. **Migration Issues**:
   ```bash
   # Reset database
   make db-reset
   
   # Run migrations manually
   go run cmd/api/main.go --migrate
   ```

3. **JWT Token Issues**:
   ```bash
   # Check JWT secret is set
   echo $JWT_SECRET
   
   # Verify token structure
   # Use jwt.io to decode tokens for debugging
   ```

### Common Frontend Issues

1. **Build Errors**:
   ```bash
   # Clear node modules and reinstall
   rm -rf node_modules package-lock.json
   npm install
   
   # Type checking
   npm run type-check
   ```

2. **API Connection Issues**:
   ```typescript
   // Check API configuration
   console.log('API URL:', config.apiUrl);
   
   // Test API endpoints manually
   curl http://localhost:8080/api/health
   ```

3. **State Management Issues**:
   ```typescript
   // Debug context state
   const MyComponent = () => {
     const auth = useAuth();
     console.log('Auth state:', auth);
     return <div>...</div>;
   };
   ```

### Development Tools

1. **Backend Debugging**:
   ```go
   // Use structured logging
   logger.Info("debug info", "user_id", userID, "action", "login")
   
   // Use Go debugger (delve)
   dlv debug cmd/api/main.go
   ```

2. **Frontend Debugging**:
   ```typescript
   // React Developer Tools
   // Browser DevTools Console
   // Network Tab for API calls
   
   // Debug API calls
   console.log('API Request:', { method, url, data });
   console.log('API Response:', response);
   ```

3. **Database Debugging**:
   ```sql
   -- Check database state
   \dt  -- List tables
   \d users  -- Describe users table
   
   -- Query logs
   SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;
   ```

---

This developer guide provides the foundation for extending and maintaining the Fullstack Template. Follow these patterns and conventions to ensure consistency and quality throughout the codebase.