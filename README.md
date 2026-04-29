# Insighta Labs+ | Backend API
Secure Profile Intelligence System

## 🏗️ System Architecture
The backend is built using a **Clean Architecture** pattern in Go.
- **cmd/api**: Entry point and route registration.
- **internals/handlers**: REST controllers for Auth and Profile logic.
- **internals/db**: Persistence layer using PostgreSQL with SQLC.
- **middlewares**: Logging, Rate Limiting, and Role-Based Access Control (RBAC).

## 🔐 Authentication Flow (OAuth + PKCE)
1. **Initiate**: Client generates a `code_challenge` and redirects to GitHub.
2. **Callback**: GitHub returns a `code` which is sent to `/auth/github/callback`.
3. **Validation**: Backend validates PKCE `code_verifier`, exchanges for GitHub token, and issues internal JWTs.
4. **Lifecycle**: 
   - Access Tokens: 3 mins
   - Refresh Tokens: 5 mins (Single-use rotation)

## 🛡️ Role Enforcement
Roles are enforced via the `RoleMiddleware`. 
- `analyst`: Read-only access to profiles.
- `admin`: Full CRUD access including profile creation and deletion.
