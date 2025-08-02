# tfa

A production-ready fullstack template with comprehensive authentication, RBAC, and user management. Built with React/TypeScript frontend and Go backend, featuring complete security, testing, and monitoring systems.

## ✨ Features

### 🔐 Complete Authentication System
- **User Registration & Login** with email verification
- **Password Management** (reset, change, strength validation)
- **JWT Authentication** with automatic token refresh
- **Session Management** with logout from all devices
- **Account Security** (rate limiting, audit trails)

### 👥 User Management
- **User Profiles** with preferences and settings
- **Role-Based Access Control (RBAC)** (user/admin roles)
- **Account Status Management** (active/inactive/suspended)
- **User Dashboard** with statistics and activity
- **Email Notifications** (welcome, verification, password reset)

### 🛡️ Security Features
- **RBAC Implementation** with role guards and protected routes
- **Input Validation** and sanitization
- **Rate Limiting** and DDoS protection
- **Security Headers** (CORS, XSS protection, etc.)
- **Audit Logging** for sensitive operations
- **SQL Injection Protection** with prepared statements

### 🏗️ Architecture
- **Clean Architecture** with domain-driven design
- **Structured Logging** with context and request tracking
- **Health Checks** and monitoring endpoints
- **Database Migrations** and seeding
- **Comprehensive Testing** (unit, integration, E2E)
- **Production Deployment** ready with Docker

### 🎯 Frontend Features
- **Modern React 18** with TypeScript and Vite
- **Responsive Design** with Tailwind CSS
- **Component Library** with reusable UI components
- **State Management** with Context API
- **Form Handling** with validation
- **Protected Routes** and role-based rendering

### 🔧 Developer Experience
- **Hot Reload** for frontend and backend
- **Comprehensive Testing** suite with >90% coverage
- **Code Quality** tools (linting, formatting)
- **API Documentation** with examples
- **Development Scripts** for common tasks
- **Docker Support** for consistent environments

## 🚀 Quick Start (< 15 minutes)

### Prerequisites

- **Go 1.23+** ([Download](https://golang.org/dl/))
- **Node.js 18+** ([Download](https://nodejs.org/))
- **PostgreSQL 13+** ([Download](https://postgresql.org/download/))
- **Git** ([Download](https://git-scm.com/downloads))

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/your-org/fullstack-template.git
cd tfa

# Install all dependencies (Go modules + npm packages)
make install
```

### 2. Database Setup

```bash
# Option A: Using Docker (Recommended)
docker-compose up -d postgres

# Option B: Local PostgreSQL
createdb fullstack_template
```

### 3. Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings (database, email, JWT secret)
# Defaults work with Docker PostgreSQL setup
```

### 4. Build and Run

```bash
# Build frontend and start development server
make frontend-build && make dev

# Or run frontend dev server separately for hot reload
make frontend-dev  # Terminal 1 (http://localhost:5173)
make dev          # Terminal 2 (API on http://localhost:8080)
```

### 5. Access the Application

- **Application**: http://localhost:8080
- **API Health**: http://localhost:8080/api/health
- **Admin Panel**: http://localhost:8080/admin (after creating admin user)

### 6. Create Admin User

```bash
# Using the built-in seed command
go run cmd/api/main.go --seed-admin

# Or register normally and promote via database:
# UPDATE users SET role = 'admin' WHERE email = 'your-email@domain.com';
```

## 📚 Core Concepts

### Authentication Flow

1. **Registration**: User creates account → email verification sent
2. **Login**: Credentials validated → JWT tokens issued
3. **Access**: Protected routes check JWT → automatic refresh
4. **Logout**: Tokens invalidated → audit log created

### Role-Based Access Control

```typescript
// Frontend: Role-based rendering
<RoleGuard requiredRole="admin">
  <AdminPanel />
</RoleGuard>

// Backend: Route protection
adminRoutes.Use(middleware.RequireRole("admin"))
```

### User Management

- **Profile Management**: Name, avatar, preferences
- **Security Settings**: Password change, session management
- **Privacy Controls**: Visibility, notification preferences
- **Audit Trail**: Track all account changes

## 🏗️ Project Structure

```
fullstack-template/
├── 📁 cmd/api/                    # Application entry point
│   └── main.go                   # Server startup and configuration
├── 📁 internal/                  # Private application code
│   ├── 📁 auth/                  # Authentication domain
│   │   ├── domain/              # Types, entities, business rules
│   │   ├── repository/          # Data access layer
│   │   ├── service/             # Business logic
│   │   └── transport/           # HTTP handlers
│   ├── 📁 user/                  # User management domain
│   ├── 📁 admin/                 # Admin operations domain
│   ├── 📁 middleware/            # HTTP middleware
│   │   ├── auth.go              # JWT authentication
│   │   ├── rbac.go              # Role-based access control
│   │   ├── rate_limit.go        # Rate limiting
│   │   └── security.go          # Security headers
│   ├── 📁 shared/               # Shared utilities
│   │   ├── config/              # Configuration management
│   │   ├── database/            # Database connection & migrations
│   │   ├── email/               # Email service (SMTP, templates)
│   │   ├── logger/              # Structured logging
│   │   └── monitoring/          # Metrics and health checks
│   └── 📁 test/integration/     # Integration tests
├── 📁 frontend/                  # React application
│   ├── 📁 src/
│   │   ├── 📁 components/       # React components
│   │   │   ├── auth/            # Authentication components
│   │   │   ├── admin/           # Admin components
│   │   │   ├── ui/              # Reusable UI components
│   │   │   └── layout/          # Layout components
│   │   ├── 📁 contexts/         # React contexts (Auth, RBAC)
│   │   ├── 📁 hooks/            # Custom React hooks
│   │   ├── 📁 lib/              # API client and utilities
│   │   ├── 📁 pages/            # Page components
│   │   ├── 📁 types/            # TypeScript type definitions
│   │   └── 📁 test/             # Frontend tests
│   ├── package.json             # NPM dependencies and scripts
│   └── vitest.config.ts         # Test configuration
├── 📁 docs/                     # Documentation
├── 🐳 docker-compose.yml        # Development environment
├── 🐳 Dockerfile               # Production container
├── ⚙️ Makefile                 # Development commands
├── 🔧 .env.example             # Environment template
└── 📖 README.md                # This file
```

## 🔧 Development Commands

### Backend Development

```bash
# Development
make dev                    # Run Go API server with hot reload
make test                   # Run all Go tests
make test-coverage          # Run tests with coverage report
make lint                   # Run Go linters (golangci-lint)

# Database
make db-migrate            # Run database migrations
make db-seed               # Seed database with test data
make db-reset              # Reset database (drop + migrate + seed)

# Building
make build                 # Build Go binary for production
make docker-build          # Build Docker image
```

### Frontend Development

```bash
# Development
make frontend-dev          # Start Vite dev server (http://localhost:5173)
make frontend-build        # Build for production
make frontend-test         # Run frontend tests
make frontend-test-ui      # Run tests with UI
make frontend-lint         # Run ESLint
make frontend-type-check   # TypeScript type checking

# Testing
cd frontend && npm run test          # Run tests
cd frontend && npm run test:coverage # Run with coverage
cd frontend && npm run test:ui       # Interactive test UI
```

### Full Stack Commands

```bash
make install               # Install all dependencies
make build-all            # Build frontend + backend
make clean                 # Clean build artifacts
make docker-dev           # Start full environment with Docker
make docker-logs          # View container logs
```

## 🧪 Testing

### Running Tests

```bash
# Backend tests
make test                  # Unit + integration tests
make test-coverage         # With coverage report

# Frontend tests  
cd frontend && npm test    # Component + API tests
cd frontend && npm run test:coverage  # With coverage

# Integration tests
make test-integration      # End-to-end testing
```

### Test Coverage

The template includes comprehensive test coverage:

- **Backend**: >90% coverage including integration tests
- **Frontend**: >85% coverage with component and API tests
- **E2E Tests**: Critical user flows and RBAC scenarios

### Test Types

1. **Unit Tests**: Individual functions and components
2. **Integration Tests**: API endpoints, database operations
3. **Component Tests**: React component behavior
4. **E2E Tests**: Complete user workflows

## 🔐 Environment Variables

### Required Variables

```bash
# Server Configuration
PORT=8080                           # Server port
ENVIRONMENT=development             # Environment (development/production)
LOG_LEVEL=info                      # Logging level

# Database
DATABASE_HOST=localhost             # PostgreSQL host
DATABASE_PORT=5432                  # PostgreSQL port
DATABASE_USER=postgres              # Database user
DATABASE_PASSWORD=postgres          # Database password
DATABASE_NAME=fullstack_template    # Database name
DATABASE_SSL_MODE=disable           # SSL mode (disable/require)

# JWT Configuration  
JWT_SECRET=your-256-bit-secret      # JWT signing secret (generate secure key)
JWT_ACCESS_DURATION=1h              # Access token lifetime
JWT_REFRESH_DURATION=720h          # Refresh token lifetime (30 days)

# Email Configuration (Optional)
EMAIL_PROVIDER=smtp                 # Email provider (smtp/mock)
SMTP_HOST=smtp.gmail.com           # SMTP server
SMTP_PORT=587                      # SMTP port
SMTP_USERNAME=your-email@gmail.com # SMTP username
SMTP_PASSWORD=your-app-password    # SMTP password
EMAIL_FROM=noreply@yourapp.com     # From email address
```

### Optional Variables

```bash
# Rate Limiting
RATE_LIMIT_REQUESTS=100            # Requests per window
RATE_LIMIT_WINDOW=1m               # Rate limit window

# Security
CORS_ORIGINS=http://localhost:3000 # Allowed CORS origins
SECURE_COOKIES=false               # Use secure cookies (true in production)

# Monitoring
METRICS_ENABLED=true               # Enable metrics collection
HEALTH_CHECK_INTERVAL=30s          # Health check interval
```

## 🚀 Deployment

### Production Deployment

1. **Build the application**:
   ```bash
   make build-all
   ```

2. **Set production environment variables**:
   ```bash
   export ENVIRONMENT=production
   export JWT_SECRET="your-production-jwt-secret"
   export DATABASE_URL="postgresql://user:pass@host:5432/dbname"
   ```

3. **Run migrations**:
   ```bash
   ./bin/api --migrate
   ```

4. **Start the server**:
   ```bash
   ./bin/api
   ```

### Docker Deployment

```bash
# Build and run with Docker
docker build -t fullstack-template .
docker run -p 8080:8080 --env-file .env fullstack-template

# Or use Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

### Deployment Checklist

- [ ] Set strong JWT secret (256-bit random key)
- [ ] Configure production database
- [ ] Set up email service (SMTP)
- [ ] Enable HTTPS/TLS
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up monitoring and logging
- [ ] Configure backup strategy
- [ ] Set proper CORS origins
- [ ] Review security headers

## 🔧 Customization & Extension

### Adding New Features

1. **Backend**: Follow the domain-driven structure
   ```bash
   internal/
   ├── newfeature/
   │   ├── domain/      # Types and business rules
   │   ├── repository/  # Data access
   │   ├── service/     # Business logic
   │   └── transport/   # HTTP handlers
   ```

2. **Frontend**: Use the component structure
   ```bash
   src/components/newfeature/
   ├── NewFeatureForm.tsx
   ├── NewFeatureList.tsx
   └── index.ts
   ```

### Extending Authentication

- **Add OAuth providers**: Extend auth service
- **Custom user fields**: Update user domain model
- **Additional roles**: Extend RBAC system
- **MFA support**: Add to auth flow

### Customizing UI

- **Theming**: Modify Tailwind config
- **Components**: Extend the UI component library
- **Layouts**: Create new layout components
- **Styling**: Use CSS modules or styled-components

## 📊 Monitoring & Health Checks

### Built-in Endpoints

- `GET /api/health` - Application health status
- `GET /api/metrics` - Prometheus metrics (if enabled)
- `GET /api/info` - Application version and environment

### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "environment": "production",
  "database": "connected",
  "email": "configured"
}
```

## 🐛 Troubleshooting

### Common Issues

**Frontend not loading**
```bash
# Ensure frontend is built
make frontend-build

# Check if dist directory exists
ls frontend/dist/

# Verify server is serving static files
curl http://localhost:8080/
```

**Database connection failed**
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify environment variables
echo $DATABASE_HOST $DATABASE_USER

# Test connection manually
psql -h localhost -U postgres -d fullstack_template
```

**Authentication not working**
```bash
# Check JWT secret is set
echo $JWT_SECRET

# Verify user is created and active
psql -c "SELECT email, status, email_verified FROM users;"

# Check browser network tab for auth errors
```

**Rate limiting errors**
```bash
# Check rate limit configuration
echo $RATE_LIMIT_REQUESTS

# Clear rate limit (Redis) or restart server
# Rate limits reset after window expires
```

### Development Tips

1. **Use Docker for consistent environment**
2. **Check logs for detailed error messages**
3. **Run tests to verify functionality**
4. **Use browser dev tools for frontend debugging**
5. **Monitor database connections and queries**

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test && cd frontend && npm test`)
6. Run linters (`make lint && make frontend-lint`)
7. Commit changes (`git commit -m 'Add amazing feature'`)
8. Push to branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code Standards

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **TypeScript/React**: Follow [React Best Practices](https://react.dev/learn)
- **Testing**: Maintain >90% backend and >85% frontend coverage
- **Documentation**: Update docs for new features
- **Security**: Follow [OWASP Guidelines](https://owasp.org/)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **API Reference**: [docs/api/](docs/api/)
- **Issues**: [GitHub Issues](https://github.com/your-org/fullstack-template/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/fullstack-template/discussions)

## 🎯 Roadmap

- [ ] **OAuth Integration** (Google, GitHub, etc.)
- [ ] **Multi-Factor Authentication (MFA)**
- [ ] **Advanced RBAC** (custom permissions)
- [ ] **Real-time Features** (WebSockets)
- [ ] **File Upload System**
- [ ] **Advanced Monitoring** (metrics, alerting)
- [ ] **API Rate Limiting per User**
- [ ] **Audit Log UI**
- [ ] **Internationalization (i18n)**
- [ ] **Progressive Web App (PWA)**

---

**Built with ❤️ for developers who value security, testing, and maintainable code.**