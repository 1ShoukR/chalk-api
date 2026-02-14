# Chalk Backend PRD

## 1) Document Metadata

- Product: `chalk-api`
- Type: Backend PRD + technical documentation
- Status: Finalized v1 (documentation baseline)
- Last updated: 2026-02-01
- Source contract: `docs/openapi.json`

## 2) Purpose

This document defines and documents the full backend system for Chalk, including architecture, modules, API scope, domain data model, integrations, and operational standards.

It is intended for:

- Backend engineering
- Frontend/mobile engineering
- QA and release validation
- Future onboarding and maintenance

## 3) System Vision

Build a Railway-first, Docker-ready, API backend for coaches and clients that supports identity, invites, workouts, messaging, scheduling, and subscriptions, with reliable side effects via transactional outbox and strong operational safety defaults.

## 4) Scope (Current Backend)

### In Scope (Implemented Foundation + MVP API)

- Auth: register, login, refresh, logout (JWT-based)
- User profile: get/update current user
- Coach profile + invite code lifecycle
- Invite preview + invite acceptance
- Workout templates and assignment + client workout execution
- Messaging conversations and message lifecycle
- Sessions scheduling and booking lifecycle
- Subscription minimum with RevenueCat webhook ingest + feature gates
- Outbox worker + event dispatcher + notification fan-out baseline
- OpenAPI contract freeze in JSON (`docs/openapi.json`)

### Out of Scope (Current Phase)

- OAuth provider sign-in flows (schema is ready, flows are not finalized in API)
- Invoice/payments full implementation
- Advanced notifications center and preference management
- Full analytics/reporting product layer

## 5) Core Architecture

### Layered Pattern (Locked)

- `Repositories` -> `Services` -> `Handlers` -> `Routes` -> `Server`
- `main.go` composes dependencies via collection initializers
- Business logic remains in services, persistence in repositories, transport in handlers

### Dependency Injection Pattern

- `InitializeRepositories` creates `RepositoriesCollection`
- `InitializeServices` creates `ServicesCollection`
- `InitializeHandlers` creates `HandlersCollection`
- `InitializeWorkers` creates `WorkersCollection`

### Transaction Strategy

- Hybrid model:
  - repository-level transactions for local operations
  - cross-domain transactional orchestration via `RepositoriesCollection.WithTransaction`
- `PublishInTx` is used for critical domain event writes to ensure atomicity with DB state changes

## 6) Runtime Bootstrap Flow

1. Configure structured logger (`slog`)
2. Load environment configuration
3. Initialize database connection (GORM + Postgres)
4. Run migrations and indexes
5. Initialize repositories
6. Initialize external integrations
7. Initialize services
8. Initialize workers (outbox processor)
9. Initialize handlers
10. Start Gin server + graceful shutdown handling

## 7) Technology Stack

### Core

- Go `1.23.1`
- Gin for HTTP transport
- GORM + Postgres driver for persistence
- Redis (`go-redis/v9`) for cache/rate-limit/session support
- `log/slog` for structured logging

### Important Dependencies

- `github.com/Netflix/go-env` for env unmarshalling
- `github.com/go-playground/validator/v10` for config/input validation
- `github.com/joho/godotenv` for local env loading
- `github.com/golang-jwt/jwt/v5` for token validation/generation
- `golang.org/x/crypto/bcrypt` for password hashing

## 8) Package Map (High-Level)

- `pkg/config`: environment model and validation
- `pkg/db`: connection bootstrap + migrations + index creation
- `pkg/models`: GORM domain models
- `pkg/repositories`: DB access and transaction helpers
- `pkg/services`: domain logic and orchestration
- `pkg/handlers`: HTTP adapters and request/response shaping
- `pkg/routes`: route registration and auth grouping
- `pkg/middleware`: auth and logging middleware
- `pkg/events`: outbox publisher, dispatcher, handlers, event types
- `pkg/workers`: outbox polling worker lifecycle
- `pkg/external`: RevenueCat, Expo, Open Food Facts integrations
- `pkg/stores`: Redis-backed stores and rate limiting helpers (fail-open)
- `pkg/utils`: shared helpers
- `pkg/errs`: custom error helpers

## 9) API Surface Summary

The canonical API contract is frozen in `docs/openapi.json`.

### Public Routes

- `GET /health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/invites/:code`
- `POST /api/v1/subscriptions/revenuecat/webhook`

### Protected Route Groups

- Auth: logout
- Users: current profile read/update
- Coaches: profile, invite codes, templates, availability, session types, coach sessions, bookable slots
- Workouts: current-user workout lifecycle, exercise completion/skip/logging
- Messages: conversations, messages, read state, unread count
- Sessions: booking + session state transitions
- Subscriptions/features: current subscription + feature access checks

## 10) Domain Modules and Capabilities

### Identity and Accounts

- User + profile records with active/banned and login tracking
- Refresh tokens and device tokens
- Password reset, email verification, magic link schemas
- JWT middleware for protected routes

### Coach and Client Relationship

- Coach profile with certifications, locations, stats
- Invite code generation, deactivation, preview, acceptance
- Client profile relationship supports one user under multiple coaches

### Workouts

- Template creation/update and exercise templating
- Assignment to clients with template deep-copy behavior
- Client workout state: start, complete, exercise-level completion/skip
- Granular set logging for workout exercises

### Messaging

- Conversation model for coach-client pair
- Message send/list/read flows
- Unread count endpoint
- `message.sent` -> `notification.push` fan-out through outbox

### Sessions

- Weekly availability, date overrides, session types
- Bookable slot computation + conflict detection
- Session lifecycle: scheduled/cancelled/completed/no_show
- Strict availability and conflict checks in booking flow

### Subscriptions

- RevenueCat webhook ingestion and normalization
- Idempotent webhook event processing
- Current subscription lookup
- Feature gate decision endpoint

## 11) Data Model Coverage

### Core Tables (By Domain)

- Users: `users`, `profiles`, `oauth_providers`, `refresh_tokens`, `device_tokens`, `password_resets`, `email_verifications`, `magic_links`
- Coach/Client: `coach_profiles`, `certifications`, `coach_locations`, `coach_stats`, `client_profiles`, `invite_codes`, `client_intake_forms`
- Workout: `workout_templates`, `workout_template_exercises`, `workouts`, `workout_exercises`, `workout_logs`
- Sessions: `coach_availabilities`, `coach_availability_overrides`, `session_types`, `sessions`
- Messaging: `conversations`, `messages`
- Subscription: `subscriptions`, `subscription_events`
- Nutrition/Progress foundation: nutrition and progress model tables are migrated for future features
- Eventing: `outbox_events`

### ID and Timestamp Strategy

- Primary keys are auto-increment integers (`uint`)
- GORM managed timestamps are retained
- No UUID primary key strategy in this backend

## 12) Eventing and Outbox Reliability

### Pattern

- Transactional outbox table persists side effects as events
- Background worker polls, claims, dispatches, retries, and marks status

### Reliability Behaviors

- `FOR UPDATE SKIP LOCKED` style claiming through repository flow
- Exponential retry backoff with attempt cap
- Permanent vs transient failure handling
- Crash recovery via requeue of stuck processing records
- Idempotency keys for dedupe-safe publishing

### Active Event Types

- `message.sent`
- `workout.assigned`
- `workout.completed`
- `session.booked`
- `invite.accepted`
- `subscription.changed`
- `notification.push`

## 13) Caching, Security Stores, and Rate Limiting

### Redis Strategy

- Redis-backed stores are initialized with fail-open behavior
- If Redis is unavailable, app favors availability over strict enforcement

### Security Limits (Current Defaults)

- Password reset: 3 attempts/hour
- Magic link: 5 attempts/hour
- Failed login tracking window: 15 minutes (monitoring, not lockout)
- Invoice creation (lenient): 10/hour per coach-client pair

### Generic Rate Limiter

- Sliding-window style increment/check helpers
- Remaining + reset-time introspection available
- Fail-open semantics preserved on Redis failure

## 14) External Integrations

### RevenueCat

- Webhook authorization via configured header value
- Event normalization + idempotent storage
- Subscription state synchronization to local model

### Expo Push

- Outbox-driven push delivery via Expo API
- Ticket error handling with retry on transient failures

### Open Food Facts

- External API client is integrated and ready
- Caching-ready store layer exists for nutrition feature expansion

## 15) Security and Compliance Posture

- JWT Bearer middleware protects private routes
- Passwords hashed with bcrypt
- Refresh-token and device-token models support session hygiene
- Sensitive values should be treated as secrets in logs/ops
- RevenueCat webhook auth enforced by configured authorization value

## 16) Operational Standards

### Logging

- Structured logging with `slog` at bootstrap and runtime boundaries

### Shutdown Behavior

- Graceful server shutdown on SIGINT/SIGTERM
- Worker stop lifecycle invoked on shutdown paths

### Migration Discipline

- Schema + index creation centralized in DB migration flow
- Conversation uniqueness and outbox indexes explicitly maintained

## 17) Environment and Deployment

### Environment Model

- Supports Railway-style `DATABASE_URL`
- Supports local split DB vars (`DB_HOST`, `DB_PORT`, etc.)
- Configurable outbox tuning params
- Configurable integration keys and webhook values

### Deployment Target

- Primary: Railway
- Packaging: Docker container
- CI baseline: GitHub Actions workflow

## 18) Quality and Testing Requirements

### Minimum Quality Gates

- Build compiles without errors
- Route contract remains aligned with `docs/openapi.json`
- Service-level authorization and ownership checks validated for core flows
- Lint/static diagnostics reviewed on touched files

### Critical End-to-End Flows to Validate

- Register -> login -> profile update
- Invite preview -> accept invite
- Coach assign workout -> client complete workout
- Session booking + cancellation/completion/no-show
- Conversation send -> unread count -> mark as read
- RevenueCat webhook -> subscription status -> feature gate response

## 19) Performance and Scalability Expectations

- Query patterns should avoid N+1 access for relation-heavy views
- Outbox processing should remain non-blocking for request/response latency
- Redis failures should not cause request hard-fail for non-critical paths
- Bookable slot and session conflict checks must remain deterministic and safe

## 20) Known Gaps and Next Backend Priorities

- Strengthen integration tests across full critical chain
- Tighten OpenAPI examples and endpoint-level schema details over time
- Expand notification strategy and preferences
- Add invoice/payments domain when product scope locks
- Deliver OAuth provider flows (Google/Facebook/Apple) post-MVP lock

## 21) Documentation Artifacts (Source of Truth)

- Backend API contract: `docs/openapi.json`
- Frontend PRD (consumer planning): `docs/frontend-prd.md`
- This backend PRD: `docs/backend-prd.md`

## 22) Final Decisions (Locked)

- Architecture remains layered with collections-based dependency initialization
- PostgreSQL + GORM are the persistence foundation
- Redis is optional but leveraged for performance and controls with fail-open policy
- Outbox pattern is mandatory for critical side effects
- Railway-first deployment remains the operational default
- API contract is managed via OpenAPI JSON and updated with backend changes
