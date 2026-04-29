# Insighta Labs+ Backend

The secure biometric intelligence core of the Insighta platform.

## System Architecture
This backend is built with Go using a modular architecture:
- **Handlers**: Process incoming HTTP requests and manage logic.
- **Middlewares**: Enforce security, rate limiting, and versioning.
- **Database**: PostgreSQL with sqlc for type-safe queries.
- **Security**: JWT-based session management with GitHub OAuth.

## Authentication Flow

### CLI Flow (PKCE)
1. User initiates `insighta login`.
2. CLI generates a PKCE `code_verifier` and `code_challenge`.
3. CLI opens a local server and directs the browser to `/auth/github/cli`.
4. User authenticates with GitHub.
5. GitHub redirects to the CLI local server.
6. CLI sends the `code` and `code_verifier` to the Backend.
7. Backend validates everything and issues tokens.

### Web Flow (HTTP-Only)
1. User clicks "Login" on the portal.
2. Web Portal fetches a secure OAuth URL from `/auth/github/url`.
3. User authenticates with GitHub.
4. GitHub redirects to `/auth/github/callback`.
5. Backend verifies the session and sets `access_token` and `refresh_token` as **HTTP-only, Secure, SameSite=Lax** cookies.

## Role Enforcement
- **Admin**: Full access to create (`POST /api/profiles`), delete (`DELETE /api/profiles/{id}`), and query profiles.
- **Analyst**: Read-only access to list, get, and search profiles.

## Rate Limiting
- **Auth Endpoints**: 10 requests per minute per User/IP.
- **API Endpoints**: 60 requests per minute per Authenticated User.

## API Versioning
All profile endpoints require the `X-API-Version: 1` header.

## Setup
1. Copy `.env.example` to `.env`.
2. Configure your GitHub OAuth Credentials.
3. Run with `go run ./cmd/api`.
