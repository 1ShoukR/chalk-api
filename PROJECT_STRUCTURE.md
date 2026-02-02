# Chalk API - Current Project Structure & Architecture

I have an existing Go API project scaffolded with the following structure and dependencies. Please regenerate the PRD to match this EXACT structure. Do not deviate from this architecture.

## Project Architecture Pattern

**Layered Architecture with Dependency Injection:**
- **Repositories** → **Services** → **Handlers** → **Routes** → **Server**
- Initialization flows from main.go through collections pattern
- Clean separation of concerns with dedicated packages

## Complete Project Structure

```
chalk-api/
├── main.go                         # Application entry point
├── go.mod / go.sum                 # Go module dependencies
├── .gitignore
├── .air.toml                       # Hot reload config (air)
├── Dockerfile                      # Production container (Railway-optimized)
├── docker-compose.yml              # Local development environment
├── railway.json                    # Railway deployment config
├── example.env                     # Environment variable template
├── Makefile                        # Development commands
│
├── .github/
│   └── workflows/
│       └── ci.yml                  # CI/CD pipeline
│
├── scripts/
│   └── run.sh                      # Shell scripts
│
└── pkg/                            # Application packages
    ├── clients/                    # External API clients (Redis, etc.)
    │   └── init_clients.go         # ClientsCollection initialization
    │
    ├── config/                     # Configuration management
    │   └── config.go               # Environment struct & loader
    │
    ├── db/                         # Database layer
    │   └── init.go                 # DB connection + RunMigrations()
    │
    ├── emails/                     # Email service
    │   └── emails.go               # Email sending logic
    │
    ├── errs/                       # Error handling
    │   └── errors.go               # Custom error types (AppError, etc.)
    │
    ├── handlers/                   # HTTP handlers (controllers)
    │   └── init_handlers.go        # HandlersCollection initialization
    │
    ├── middleware/                 # HTTP middleware
    │   ├── auth.go                 # Auth middleware (JWT - to be implemented)
    │   └── logger.go               # Structured logging setup
    │
    ├── models/                     # Database models (GORM)
    │   └── base.go                 # BaseModel with common fields
    │
    ├── repositories/               # Data access layer
    │   └── init_repositories.go    # RepositoriesCollection initialization
    │
    ├── routes/                     # Route definitions
    │   └── routes.go               # Gin router setup
    │
    ├── seeds/                      # Database seeding
    │   └── seeds.go                # Seed data initialization
    │
    ├── server/                     # HTTP server
    │   └── server.go               # Server struct with Start/Shutdown
    │
    ├── services/                   # Business logic layer
    │   └── init_services.go        # ServicesCollection initialization
    │
    ├── stores/                     # Runtime stores (caches)
    │   └── stores.go               # StoresCollection (Redis-backed)
    │
    ├── utils/                      # Helper functions
    │   └── helpers.go              # String utils, slugify, etc.
    │
    ├── version/                    # Version information
    │   └── version.go              # Build version/commit/date
    │
    └── workers/                    # Background workers
        └── init_workers.go         # WorkersCollection initialization
```

## Dependencies (go.mod)

### Direct Dependencies

1. **github.com/Netflix/go-env** (v0.1.2)
   - Environment variable parsing into structs
   - Used in: `pkg/config/config.go`

2. **github.com/gin-gonic/gin** (v1.10.0)
   - HTTP web framework for REST APIs
   - Used in: `pkg/server/`, `pkg/routes/`, `pkg/middleware/`

3. **github.com/go-playground/validator/v10** (v10.24.0)
   - Struct field validation
   - Used in: `pkg/config/config.go`

4. **github.com/google/uuid** (v1.6.0)
   - UUID generation for records
   - Used in: `pkg/models/base.go`

5. **github.com/joho/godotenv** (v1.5.1)
   - Loads .env files for local development
   - Used in: `pkg/config/config.go`

6. **gorm.io/driver/postgres** (v1.5.11)
   - PostgreSQL driver for GORM
   - Used in: `pkg/db/init.go`

7. **gorm.io/gorm** (v1.25.12)
   - ORM for database operations
   - Used throughout: `pkg/db/`, `pkg/models/`, `pkg/repositories/`

## Initialization Flow (main.go)

```
1. Setup Logging (slog)
2. Load Config (Environment struct)
3. Initialize Database (GORM + PostgreSQL)
4. Run Migrations (db.RunMigrations)
5. Initialize Repositories (RepositoriesCollection)
6. Initialize Services (ServicesCollection)
7. Initialize Handlers (HandlersCollection)
8. Create Server (Gin)
9. Start Server with graceful shutdown
```

## Environment Variables (example.env)

```env
# Server
PORT=8080
RUN_MODE=local

# Database (Railway provides DATABASE_URL automatically)
DATABASE_URL=postgresql://...       # Railway style
# OR individual vars for local:
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=chalkdb

# Redis (optional)
REDIS_URL=localhost:6379

# JWT Authentication (to be implemented by me)
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

# OAuth (to be implemented by me)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
FACEBOOK_CLIENT_ID=
FACEBOOK_CLIENT_SECRET=
APPLE_CLIENT_ID=
APPLE_TEAM_ID=
APPLE_KEY_ID=
```

## Deployment

- **Platform**: Railway
- **Database**: PostgreSQL (Railway-managed)
- **Container**: Docker (distroless base image)
- **CI/CD**: GitHub Actions

## Important Notes for PRD

1. **DO NOT use UUIDs** - I will implement my own ID strategy
2. **DO NOT implement JWT/OAuth** - I will build authentication myself
3. **DO NOT use Firebase** - Not using Firebase for anything
4. **Keep this exact folder structure** - All packages in `pkg/` directory
5. **Use Collections pattern** - All init files return a `*Collection` struct
6. **Railway-first** - Config supports both `DATABASE_URL` and individual DB vars
7. **GORM for ORM** - All database operations through GORM
8. **Gin for HTTP** - All routes and handlers use Gin framework

## Architecture Principles

- **Clean layered architecture**: Handlers → Services → Repositories → DB
- **Dependency injection**: Collections passed down the initialization chain
- **Environment-based config**: `.env` for local, env vars for production
- **Graceful shutdown**: Signal handling for clean server termination
- **Structured logging**: `log/slog` throughout the application
- **Docker-optimized**: Multi-stage builds, distroless runtime

Please regenerate the PRD based on this EXACT structure and dependencies.
