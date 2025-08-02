# Fullstack Template - Development Guide

This document contains development information and commands for working with the fullstack template.

## Quick Commands

### Development
```bash
# Install all dependencies
make install

# Start development (builds frontend + runs Go server)
make frontend-build && make dev

# Frontend development server (separate from Go API)
make frontend-dev

# Run Go API only
make dev
```

### Building
```bash
# Build frontend for production
make frontend-build

# Build Go binary
make build

# Build everything
make build-all
```

### Testing & Quality
```bash
# Run Go tests
make test

# Run Go tests with coverage
make test-coverage

# Lint Go code
make lint

# Lint frontend code
make frontend-lint

# Type check frontend
cd frontend && npm run type-check
```

### Docker
```bash
# Build Docker image
make docker-build

# Run with docker-compose (includes PostgreSQL)
docker-compose up

# Run container directly
make docker-run
```

## Project Architecture

### Backend (Go)
- **Clean Architecture**: Domain/Service/Transport layers
- **Gin Framework**: HTTP router and middleware
- **GORM**: Database ORM for PostgreSQL
- **Structured Logging**: Using slog
- **Health Checks**: Built-in `/api/health` endpoint

### Frontend (React)
- **React 18**: With TypeScript
- **Vite**: Fast development and building
- **Tailwind CSS**: Utility-first styling
- **Component Architecture**: Layout/Sections/UI separation

### Deployment
- **Single Binary**: Go server serves both API and frontend
- **Static Assets**: Frontend built to `/frontend/dist/`
- **SPA Routing**: All non-API routes serve `index.html`

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Server
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info

# Database (optional for development)
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=fullstack_template
```

## Common Development Tasks

### Adding New API Endpoints
1. Create domain types in `internal/{module}/domain/`
2. Implement business logic in `internal/{module}/service/`
3. Add HTTP handlers in `internal/{module}/transport/`
4. Register routes in `internal/http/server.go`

### Adding New React Components
1. Create components in appropriate directory:
   - `frontend/src/components/ui/` - Reusable UI components
   - `frontend/src/components/layout/` - Layout components
   - `frontend/src/components/sections/` - Page sections
2. Export from `index.ts` files for clean imports
3. Use TypeScript interfaces for props

### Database Changes
1. Update domain models in `internal/{module}/domain/`
2. Run migrations (implement as needed)
3. Update repository layer if needed

## Troubleshooting

### Frontend not loading
- Ensure `make frontend-build` has been run
- Check that `/frontend/dist/` directory exists
- Verify Go server is serving static files correctly

### Database connection issues
- Check environment variables
- Ensure PostgreSQL is running (use `docker-compose up postgres`)
- Verify connection string format

### Build issues
- Run `go mod tidy` to update dependencies
- Clear frontend node_modules: `rm -rf frontend/node_modules && npm install`
- Check Go and Node versions match requirements

## File Structure Reference

```
.
├── cmd/api/main.go           # Application entry point
├── internal/                 # Private Go code
│   ├── health/              # Health check module
│   ├── info/                # App info module  
│   ├── http/                # HTTP server setup
│   ├── middleware/          # HTTP middleware
│   └── shared/              # Shared utilities
├── frontend/                # React application
│   ├── src/components/      # React components
│   ├── src/config/          # Frontend configuration
│   └── public/              # Static assets
├── Dockerfile               # Multi-stage build
├── docker-compose.yml       # Development with PostgreSQL
└── Makefile                # Development commands
```

This template follows patterns from nimbus and api-template projects for consistency and best practices.