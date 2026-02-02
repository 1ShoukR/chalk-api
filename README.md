# chalk-api

RESTful API backend for the Chalk personal training platform.

## Tech Stack

- **Language:** Go 1.23+
- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL (Railway)
- **Cache:** Redis (Railway)
- **Deployment:** Railway
- **Storage:** Cloudflare R2
- **Subscriptions:** RevenueCat

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose (for local development)

### Local Development

1. **Copy environment variables:**
   ```bash
   cp example.env .env
   ```

2. **Start database and Redis:**
   ```bash
   make docker-up
   ```

3. **Run the API:**
   ```bash
   make run
   ```

4. **Access:**
   - API: http://localhost:8080
   - Health check: http://localhost:8080/health

### Available Commands

```bash
make run          # Run API locally
make dev          # Run with hot reload (requires air)
make docker-up    # Start Docker services
make docker-down  # Stop Docker services
make test         # Run tests
make build        # Build binary
```

## Project Structure

```
chalk-api/
├── main.go                 # Application entry point
├── pkg/
│   ├── models/            # Database models
│   ├── repositories/      # Data access layer
│   ├── services/          # Business logic
│   ├── handlers/          # HTTP handlers
│   ├── routes/            # Route definitions
│   ├── middleware/        # HTTP middleware
│   ├── config/            # Configuration
│   └── db/                # Database setup
└── ...
```

## Database Models

- **User System:** User, Profile, OAuth, Tokens
- **Coach System:** CoachProfile, Certifications, Locations, Stats
- **Client System:** ClientProfile, InviteCode, IntakeForm
- **Subscriptions:** Subscription, SubscriptionEvent
- _(More to come: Exercise, Workout, Template, Session, Nutrition, Progress, Messages)_

## License

Proprietary
