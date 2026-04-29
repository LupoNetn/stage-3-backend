# Insighta Labs+ | Backend API
Secure Profile Intelligence System

**Live API**: [https://stage-3-backend-azure.vercel.app](https://stage-3-backend-azure.vercel.app)

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
Roles are enforced via the `RoleMiddleware` and `Authorize` higher-order functions.
- `analyst`: Read-only access to `/api/profiles`.
- `admin`: Full CRUD access including profile creation and deletion.

## 🔎 Natural Language Search Approach
The system uses a **Rule-Based Tokenizer** combined with a **Vectorized SQL Query Builder**.
- **Tokenization**: Filters noise and extracts key entities (Genders, Countries, Age Groups).
- **Interpretable Querying**: Translates natural language into structured SQL `WHERE` clauses for high-precision filtering across thousands of profiles.
