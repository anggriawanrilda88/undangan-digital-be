# Undangan Digital — Golang API

REST API backend untuk platform Web Undangan Digital Pernikahan.

## Tech Stack

- **Framework:** Gin
- **ORM:** GORM
- **Database:** PostgreSQL (via Supabase)
- **Auth:** Supabase JWT verification

## Project Structure

```
undangan-api/
├── main.go
├── go.mod
├── .env.example
├── app/
│   ├── dto/              # Request & Response DTOs
│   └── usecase/          # Business logic
├── domain/
│   ├── entities/         # Core domain entities
│   ├── errors/           # Domain errors
│   └── repositories/     # Repository interfaces
├── infra/
│   ├── models/           # GORM models
│   └── persistence/      # Repository implementations + DB connection
├── interface/
│   └── rest/v1/
│       ├── router.go     # Route registration
│       ├── middleware/   # Auth middleware
│       ├── response/     # Standard response helpers
│       ├── invitation/   # Invitation handlers
│       └── rsvp/         # RSVP handlers
└── migration/            # SQL migration files
```

## Setup

```bash
# 1. Copy env
cp .env.example .env
# Edit .env — isi DATABASE_URL dan SUPABASE_JWT_SECRET

# 2. Install dependencies
go mod tidy

# 3. Run
go run main.go
```

## API Endpoints

### Public (no auth)
```
GET  /api/v1/i/:slug              # Halaman undangan publik
POST /api/v1/invitations/:id/rsvp # Submit RSVP tamu
```

### Authenticated (Bearer JWT)
```
GET  /api/v1/auth/me

GET  /api/v1/invitations
POST /api/v1/invitations
GET  /api/v1/invitations/:id
PUT  /api/v1/invitations/:id
DEL  /api/v1/invitations/:id
GET  /api/v1/invitations/:id/rsvp

GET  /api/v1/slugs/check?slug=xxx  ← US-04 blocker
```

## Auth

Semua endpoint protected menggunakan Supabase JWT.  
Header: `Authorization: Bearer <supabase_access_token>`

## OpenAPI Spec

Lihat: `../docs/openapi-undangan-digital.yaml`  
Preview: https://editor.swagger.io
