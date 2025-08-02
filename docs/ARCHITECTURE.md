# Architecture Documentation

This document describes the system architecture, design patterns, and architectural decisions for the Fullstack Template.

## Table of Contents

- [System Overview](#system-overview)
- [Backend Architecture](#backend-architecture)
- [Frontend Architecture](#frontend-architecture)
- [Security Architecture](#security-architecture)
- [Data Architecture](#data-architecture)
- [Infrastructure Architecture](#infrastructure-architecture)
- [Design Patterns](#design-patterns)
- [Architectural Decisions](#architectural-decisions)

---

## System Overview

The Fullstack Template follows a **Clean Architecture** approach with clear separation of concerns, dependency inversion, and domain-driven design principles.

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (React/TS)    │◄──►│   (Go/Gin)      │◄──►│  (PostgreSQL)   │
│                 │    │                 │    │                 │
│ • Components    │    │ • Clean Arch    │    │ • Users         │
│ • State Mgmt    │    │ • RBAC          │    │ • Audit Logs    │
│ • Auth Context  │    │ • JWT Auth      │    │ • Sessions      │
│ • API Client    │    │ • Rate Limiting │    │ • Preferences   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         └─────────────►│   Static Files  │◄─────────────┘
                        │   (Embedded)    │
                        └─────────────────┘
```

### Key Architectural Principles

1. **Separation of Concerns**: Clear boundaries between layers
2. **Dependency Inversion**: High-level modules don't depend on low-level modules
3. **Single Responsibility**: Each component has one reason to change
4. **Domain-Driven Design**: Business logic isolated in domain layer
5. **Security by Design**: Security considerations at every layer
6. **Testability**: Architecture supports comprehensive testing

---

## Backend Architecture

The backend follows **Clean Architecture** with domain-driven design, organized into layers with clear dependencies.

### Layer Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                        Transport Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ HTTP        │  │ Middleware  │  │ WebSocket   │             │
│  │ Handlers    │  │ (Auth/RBAC) │  │ (Future)    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Service Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Auth        │  │ User        │  │ Admin       │             │
│  │ Service     │  │ Service     │  │ Service     │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Domain Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ User        │  │ Auth        │  │ Audit       │             │
│  │ Entities    │  │ Entities    │  │ Entities    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Repository Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ User        │  │ Auth        │  │ Audit       │             │
│  │ Repository  │  │ Repository  │  │ Repository  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Database    │  │ Email       │  │ Logger      │             │
│  │ (GORM)      │  │ Service     │  │ (slog)      │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

### Module Structure

Each domain module follows a consistent structure:

```
internal/
├── auth/                    # Authentication domain
│   ├── domain/             # Domain entities and business rules
│   │   ├── types.go        # Core types and entities
│   │   ├── errors.go       # Domain-specific errors
│   │   └── rbac.go         # RBAC definitions
│   ├── repository/         # Data access layer
│   │   ├── user.go         # User data operations
│   │   ├── refresh_token.go # Token management
│   │   └── password_reset.go # Password reset operations
│   ├── service/            # Business logic layer
│   │   ├── auth.go         # Authentication service
│   │   ├── jwt.go          # JWT token service
│   │   └── email.go        # Email service
│   └── transport/          # HTTP transport layer
│       └── http.go         # HTTP handlers
```

### Dependency Flow

```
Transport → Service → Repository → Database
    ↓         ↓         ↓           ↓
  HTTP    Business   Data      Storage
 Layer     Logic    Access      Layer
```

**Key Rules:**
- Transport layer depends on Service layer
- Service layer depends on Repository interfaces (not implementations)
- Repository layer implements interfaces defined in Service layer
- Domain layer has no external dependencies

### Service Layer Design

Services encapsulate business logic and coordinate between repositories:

```go
type AuthService struct {
    config            *config.Config
    logger            *slog.Logger
    userRepo          UserRepository      // Interface
    refreshTokenRepo  RefreshTokenRepository
    passwordResetRepo PasswordResetRepository
    jwtService        *JWTService
    emailService      *EmailService
}

// Business logic methods
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error)
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error)
func (s *AuthService) RefreshToken(req *RefreshTokenRequest) (*AuthResponse, error)
```

### Repository Pattern

Repositories abstract data access and provide a clean interface:

```go
type UserRepository interface {
    Create(user *User) error
    GetByID(id uint) (*User, error)
    GetByEmail(email string) (*User, error)
    Update(user *User) error
    Delete(id uint) error
    ExistsByEmail(email string) (bool, error)
}

type userRepository struct {
    db *gorm.DB
}

func (r *userRepository) Create(user *User) error {
    return r.db.Create(user).Error
}
```

---

## Frontend Architecture

The frontend uses **Component-Based Architecture** with React, emphasizing reusability, maintainability, and type safety.

### Component Hierarchy

```
App
├── AuthProvider                 # Authentication context
│   ├── RBACProvider            # Role-based access context
│   │   ├── Router              # Route management
│   │   │   ├── PublicRoutes    # Non-authenticated routes
│   │   │   │   ├── LoginPage
│   │   │   │   ├── RegisterPage
│   │   │   │   └── ForgotPasswordPage
│   │   │   └── ProtectedRoutes # Authenticated routes
│   │   │       ├── Dashboard
│   │   │       ├── Profile
│   │   │       └── AdminPanel (RoleGuard: admin)
│   │   └── Layout              # Common layout
│   │       ├── Header
│   │       ├── Navigation
│   │       └── Footer
│   └── Components              # Reusable components
│       ├── UI Components       # Generic UI elements
│       ├── Auth Components     # Authentication forms
│       └── Domain Components   # Business-specific components
```

### State Management Architecture

```
┌─────────────────┐
│   AuthContext   │     Global authentication state
│                 │     • User data
│                 │     • Authentication status
│                 │     • Token management
└─────────────────┘
          │
┌─────────────────┐
│   RBACContext   │     Role-based access control
│                 │     • Permission checks
│                 │     • Role validation
│                 │     • Access control
└─────────────────┘
          │
┌─────────────────┐
│ Component State │     Local component state
│                 │     • Form data
│                 │     • UI state
│                 │     • Temporary data
└─────────────────┘
```

### Component Organization

```
src/
├── components/
│   ├── ui/                  # Generic, reusable components
│   │   ├── Button.tsx       # Base button component
│   │   ├── Input.tsx        # Form input component
│   │   ├── Modal.tsx        # Modal component
│   │   └── index.ts         # Barrel exports
│   ├── auth/                # Authentication components
│   │   ├── LoginForm.tsx    # Login form
│   │   ├── RegisterForm.tsx # Registration form
│   │   ├── ProtectedRoute.tsx # Route protection
│   │   ├── RoleGuard.tsx    # Role-based rendering
│   │   └── index.ts
│   ├── layout/              # Layout components
│   │   ├── Header.tsx
│   │   ├── Footer.tsx
│   │   ├── Navigation.tsx
│   │   └── index.ts
│   └── admin/               # Admin-specific components
│       ├── UserList.tsx
│       ├── UserModal.tsx
│       └── index.ts
├── contexts/                # React contexts
│   ├── AuthContext.tsx      # Authentication context
│   └── RBACContext.tsx      # RBAC context
├── hooks/                   # Custom hooks
│   ├── useAuth.ts           # Authentication hook
│   ├── useRBAC.ts           # RBAC hook
│   └── useApi.ts            # API hook
├── lib/                     # Utilities and services
│   ├── api.ts               # API client
│   ├── auth.ts              # Auth utilities
│   └── validation.ts        # Form validation
├── pages/                   # Page components
│   ├── Dashboard.tsx
│   ├── Profile.tsx
│   └── admin/
│       └── AdminPanel.tsx
└── types/                   # TypeScript definitions
    ├── api.ts               # API types
    ├── auth.ts              # Auth types
    └── user.ts              # User types
```

### API Client Architecture

The API client provides a centralized interface for backend communication:

```typescript
class ApiClient {
  private baseURL: string;
  private refreshPromise: Promise<void> | null = null;

  // Automatic token refresh on 401 errors
  private async request<T>(endpoint: string, options: RequestInit): Promise<T> {
    let response = await fetch(url, config);
    
    // Handle token refresh on 401 errors
    if (response.status === 401 && endpoint !== '/auth/refresh') {
      const refreshed = await this.refreshToken();
      if (refreshed) {
        response = await fetch(url, config); // Retry request
      }
    }
    
    return this.handleResponse(response);
  }

  // Authentication methods
  async login(data: LoginRequest): Promise<AuthResponse> { }
  async register(data: RegisterRequest): Promise<AuthResponse> { }
  async refreshToken(): Promise<boolean> { }
}
```

---

## Security Architecture

Security is implemented at multiple layers with defense-in-depth principles.

### Authentication & Authorization Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │  Middleware │    │   Service   │    │  Database   │
│             │    │             │    │             │    │             │
│ 1. Login    │───►│ 2. Validate │───►│ 3. Verify   │───►│ 4. Check    │
│ Request     │    │ Input       │    │ Credentials │    │ User        │
│             │    │             │    │             │    │             │
│ 8. Store    │◄───│ 7. Return   │◄───│ 6. Generate │    │ 5. Return   │
│ Tokens      │    │ Tokens      │    │ JWT Tokens  │    │ User Data   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Security Layers

1. **Transport Security**
   - HTTPS in production
   - Security headers (HSTS, CSP, etc.)
   - CORS configuration

2. **Authentication Security**
   - JWT tokens with short expiration
   - Secure refresh token rotation
   - Password hashing with bcrypt
   - Rate limiting on auth endpoints

3. **Authorization Security**
   - Role-Based Access Control (RBAC)
   - Route-level protection
   - Component-level protection
   - API endpoint protection

4. **Input Security**
   - Input validation and sanitization
   - SQL injection prevention
   - XSS protection
   - CSRF protection

5. **Session Security**
   - Secure token storage
   - Automatic token refresh
   - Session invalidation
   - Device tracking

### RBAC Implementation

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    User     │    │    Role     │    │ Permission  │
│             │    │             │    │             │
│ • ID        │───►│ • user      │───►│ • read_own  │
│ • Email     │    │ • admin     │    │ • write_own │
│ • Role      │    │             │    │ • read_all  │
│ • Status    │    │             │    │ • write_all │
└─────────────┘    └─────────────┘    └─────────────┘
```

**Role Hierarchy:**
- `user`: Basic user permissions (read/write own data)
- `admin`: Full system access (read/write all data)

**Implementation:**
```go
// Backend middleware
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetUserFromContext(c)
        if !user.HasRole(role) {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Frontend component guard
<RoleGuard requiredRole="admin">
  <AdminPanel />
</RoleGuard>
```

---

## Data Architecture

The data layer uses PostgreSQL with GORM ORM, following database design best practices.

### Database Schema

```sql
-- Users table (core entity)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verify_token VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    preferences JSONB DEFAULT '{}',
    avatar VARCHAR(255),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP -- Soft delete
);

-- Refresh tokens (session management)
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Password reset tokens
CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Audit logs (security and compliance)
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    target_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    level VARCHAR(20) DEFAULT 'info',
    resource VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Data Relationships

```
Users (1) ──────────── (N) Refresh Tokens
  │
  │ (as user)
  └─── (N) Audit Logs
  │
  │ (as target)
  └─── (N) Audit Logs

Password Resets (N) ──────── (1) Users (by email)
```

### Data Access Patterns

1. **Repository Pattern**: Abstracts data access
2. **Unit of Work**: Manages transactions
3. **Soft Deletes**: Maintains data integrity
4. **Audit Trails**: Tracks all changes
5. **JSONB Storage**: Flexible user preferences

### Database Migrations

```go
// Auto-migration on startup
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &User{},
        &RefreshToken{},
        &PasswordReset{},
        &AuditLog{},
    )
}

// Manual migrations for production
func RunMigrations(db *gorm.DB) error {
    // Create tables
    // Add indexes
    // Update constraints
    return nil
}
```

---

## Infrastructure Architecture

The infrastructure supports both development and production environments.

### Development Environment

```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │
│   (Vite Dev)    │    │   (Go Direct)   │
│   :5173         │    │   :8080         │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┼─────────────────┐
                                 │                 │
                    ┌─────────────────┐    ┌─────────────────┐
                    │   PostgreSQL    │    │   Mail Catcher  │
                    │   (Docker)      │    │   (Optional)    │
                    │   :5432         │    │   :1025         │
                    └─────────────────┘    └─────────────────┘
```

### Production Environment

```
┌─────────────────┐
│   Load Balancer │
│   (nginx/haproxy)│
│   :80, :443     │
└─────────────────┘
         │
┌─────────────────┐
│   Go Binary     │
│   (Static +     │
│    API Server)  │
│   :8080         │
└─────────────────┘
         │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   SMTP Server   │    │   Monitoring    │
│   (Managed)     │    │   (SendGrid/    │    │   (Prometheus)  │
│   :5432         │    │    AWS SES)     │    │   :9090         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Container Architecture

```dockerfile
# Multi-stage build
FROM node:18-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN go build -o bin/api cmd/api/main.go

FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=backend /app/bin/api ./
EXPOSE 8080
CMD ["./api"]
```

### Monitoring & Observability

1. **Health Checks**: `/health` endpoint for load balancer
2. **Structured Logging**: JSON logs with correlation IDs
3. **Metrics**: Prometheus metrics for monitoring
4. **Distributed Tracing**: Request tracing (optional)
5. **Audit Logging**: Security and compliance tracking

---

## Design Patterns

### Backend Patterns

1. **Clean Architecture**
   - Dependency Inversion
   - Separation of Concerns
   - Testable Architecture

2. **Repository Pattern**
   - Data Access Abstraction
   - Testable Data Layer
   - Interface Segregation

3. **Service Layer Pattern**
   - Business Logic Encapsulation
   - Transaction Management
   - Dependency Coordination

4. **Middleware Pattern**
   - Cross-cutting Concerns
   - Request/Response Processing
   - Plugin Architecture

5. **Strategy Pattern**
   - Email Provider Selection
   - Authentication Methods
   - Configurable Behaviors

### Frontend Patterns

1. **Component Composition**
   - Reusable Components
   - Props-based Configuration
   - Children Composition

2. **Context Provider Pattern**
   - Global State Management
   - Dependency Injection
   - Cross-component Communication

3. **Higher-Order Components**
   - Cross-cutting Concerns
   - Behavior Enhancement
   - Code Reuse

4. **Custom Hooks Pattern**
   - Stateful Logic Reuse
   - Side Effect Management
   - API Integration

5. **Render Props Pattern**
   - Dynamic Rendering
   - Logic Sharing
   - Component Enhancement

---

## Architectural Decisions

### ADR-001: Clean Architecture

**Status**: Accepted

**Context**: Need scalable, maintainable backend architecture

**Decision**: Implement Clean Architecture with domain-driven design

**Consequences**:
- ✅ Clear separation of concerns
- ✅ Testable architecture
- ✅ Technology independence
- ❌ Initial complexity
- ❌ More boilerplate code

### ADR-002: JWT Authentication

**Status**: Accepted

**Context**: Need stateless authentication for scalability

**Decision**: Use JWT tokens with refresh token rotation

**Consequences**:
- ✅ Stateless authentication
- ✅ Scalable across instances
- ✅ Mobile-friendly
- ❌ Token size overhead
- ❌ Revocation complexity

### ADR-003: PostgreSQL Database

**Status**: Accepted

**Context**: Need reliable, ACID-compliant database

**Decision**: Use PostgreSQL with GORM ORM

**Consequences**:
- ✅ ACID compliance
- ✅ Rich feature set
- ✅ JSON support
- ✅ Excellent tooling
- ❌ Requires setup/maintenance

### ADR-004: React Frontend

**Status**: Accepted

**Context**: Need modern, maintainable frontend

**Decision**: Use React with TypeScript and Vite

**Consequences**:
- ✅ Component-based architecture
- ✅ Large ecosystem
- ✅ Type safety
- ✅ Fast development
- ❌ Bundle size
- ❌ Learning curve

### ADR-005: Embedded Static Files

**Status**: Accepted

**Context**: Simplify deployment with single binary

**Decision**: Embed frontend files in Go binary

**Consequences**:
- ✅ Single binary deployment
- ✅ Simplified operations
- ✅ No separate web server needed
- ❌ Larger binary size
- ❌ Rebuild required for frontend changes

### ADR-006: RBAC Implementation

**Status**: Accepted

**Context**: Need authorization system

**Decision**: Implement simple role-based access control

**Consequences**:
- ✅ Simple to understand
- ✅ Sufficient for most use cases
- ✅ Easy to implement
- ❌ Less flexible than ABAC
- ❌ May need extension for complex cases

---

## Performance Considerations

### Backend Performance

1. **Database Optimization**
   - Proper indexing strategy
   - Query optimization
   - Connection pooling
   - Prepared statements

2. **Caching Strategy**
   - In-memory caching for static data
   - Redis for session storage (optional)
   - HTTP caching headers

3. **Concurrency**
   - Go goroutines for concurrent processing
   - Context-based cancellation
   - Proper mutex usage

### Frontend Performance

1. **Code Splitting**
   - Route-based splitting
   - Component lazy loading
   - Dynamic imports

2. **Bundle Optimization**
   - Tree shaking
   - Minification
   - Compression

3. **Rendering Optimization**
   - React.memo for components
   - useMemo/useCallback for expensive operations
   - Virtual scrolling for large lists

### Scalability Patterns

1. **Horizontal Scaling**
   - Stateless application design
   - Load balancer support
   - Database connection management

2. **Vertical Scaling**
   - Efficient resource usage
   - Memory management
   - CPU optimization

3. **Caching Layers**
   - Application-level caching
   - Database query caching
   - CDN for static assets

---

This architecture provides a solid foundation for building scalable, maintainable applications while maintaining security and performance standards. The modular design allows for easy extension and modification as requirements evolve.