# HNG Stage 1 - Personal Profile Management API

A robust Go-based RESTful API that aggregates personal predictive data from external sources, provides persistent storage, and serves filtered profiles.

## 🚀 Features

- **Concurrent Data Aggregation**: Uses `errgroup` to fetch predictive data from Genderize, Agify, and Nationalize APIs simultaneously.
- **Smart Data Processing**:
  - Automatically classifies age into groups (child, teenager, adult, senior).
  - Selects the country with the highest probability.
  - Generates **UUID v7** identifiers for new profiles.
- **Database Persistence**: Fully integrated with PostgreSQL using `pgx` and `sqlc` for high-performance data access.
- **Filtered Retrieval**: Supports case-insensitive query parameters for filtering profiles by gender, country, and age group.
- **Idempotent Creation**: Prevents duplicate profile creation if the same name is provided again.
- **Modern Standards**: Clean code architecture with separated concerns (handlers, types, utils, db).

## 🛠️ Technology Stack

- **Language**: Go (Golang)
- **Database**: PostgreSQL (via [Neon](https://neon.tech/))
- **Drivers/Tools**:
  - [pgx](https://github.com/jackc/pgx): PostgreSQL driver and toolkit.
  - [sqlc](https://sqlc.dev/): Type-safe SQL generator.
  - [uuid](https://github.com/google/uuid): UUID v7 generation.
  - [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup): Concurrent execution control.
  - [Air](https://github.com/cosmtrek/air): Live reloading for development.

## 🏁 Getting Started

### Prerequisites

- Go 1.25.5 or higher
- A running PostgreSQL database (or Neon connection string)

### Setup

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd stage-1
   ```

2. **Configure environment variables**:
   Create a `.env` file in the root directory:
   ```env
   PORT=7000
   DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

### Running Locally

You can run the application normally:
```bash
go run ./cmd/api/main.go
```

Or use **Air** for live-reloading during development:
```bash
air
```

## 📖 API Documentation

### Create Profile
`POST /api/profiles`
- **Body**: `{"name": "ella"}`
- **Success (201)**: Returns the processed profile.
- **Idempotency**: If the profile already exists, returns the existing one with a message.

### Get All Profiles (with filtering)
`GET /api/profiles`
- **Optional Params**: `gender`, `country_id`, `age_group` (all are case-insensitive).
- **Example**: `/api/profiles?gender=female&country_id=DRC`

### Get Profile by ID
`GET /api/profiles/{id}`

### Delete Profile
`DELETE /api/profiles/{id}`
- **Success (204)**: No Content.

## 🧪 Testing

You can use `curl` to quickly test the endpoints:
```bash
# Create Profile
curl -X POST http://localhost:7000/api/profiles -H "Content-Type: application/json" -d '{"name": "ella"}'

# Get All Profiles
curl http://localhost:7000/api/profiles
```

## 🚀 Deployment

This project is configured for deployment on **Vercel**. Ensure you add your `DATABASE_URL` to Vercel's Environment Variables in the project settings.


## 🔍 Natural Language Search (Core Feature)

The system includes a custom, rule-based Natural Language Query (NLQ) engine that allows users to search for profiles using plain English queries via the `GET /api/profiles/search?q=...` endpoint.

### 1. Parsing Approach
The parser is built as a **deterministic, rule-based engine** (no LLMs or external AI dependencies). The logic follows these steps:
1.  **Normalization**: The query is lowercased and stripped of special characters using regex.
2.  **Tokenization**: The query is split into individual words (tokens).
3.  **Keyword Mapping**: The engine scans tokens for specific keywords and maps them to database filters:
    *   **Gender**: Maps `man`, `men`, `boy`, `male` → `male`; `woman`, `women`, `girl`, `female` → `female`. Supports "male and female" combinations.
    *   **Age Ranges**: Interprets numeric tokens following comparison keywords:
        *   `above [age]`, `older than [age]`, `not younger than [age]` → sets `min_age`.
        *   `below [age]`, `younger than [age]`, `not older than [age]` → sets `max_age`.
    *   **Special Mapping**: The keyword `young` is explicitly mapped to an age range of **16–24** per project specifications.
    *   **Location Parsing**: The engine looks for `from` or `in`. The token immediately following is treated as a location:
        *   If the token length is < 3 (e.g., `NG`, `US`), it is mapped to `country_id`.
        *   Otherwise (e.g., `Nigeria`, `Kenya`), it is mapped to `country_name`.
4.  **Interpretability Validation**: Every query is checked for "interpretability." If no keywords are recognized that map to a filter, the system returns a `422 Unprocessable Entity` error with the message "Unable to interpret query."

### 2. Supported Keywords
| Category | Keywords | Example Query |
| :--- | :--- | :--- |
| **Gender** | `male`, `female`, `man`, `woman`, `boy`, `girl` | `young males` |
| **Age Group** | `teenager`, `adult`, `senior`, `child` | `adults from Kenya` |
| **Comparison** | `above`, `below`, `older than`, `younger than` | `females above 30` |
| **Location** | `from`, `in` | `people from Nigeria` |
| **Range Keyword** | `young` (maps to 16-24) | `young females` |

### 3. Limitations & Edge Cases
*   **Multi-word Locations**: The current parser captures only the single word immediately following "from" or "in". It does not currently support multi-word country names like `South Africa` or `United Kingdom`.
*   **Boolean Complexity**: The parser supports simple combinations (e.g., "male and female") but does not handle complex nested logic like `(males from Kenya) OR (females from Nigeria)`.
*   **Typo Tolerance**: The system requires exact keyword matches. It does not handle typos or fuzzy matching (e.g., "femail" will not be recognized as "female").
*   **Contextual Ambiguity**: If a number is provided without a comparison keyword (e.g., "females 30"), the system treats it as an exact age match.
*   **Database Integration**: All filtering is performed directly in the PostgreSQL database for performance, ensuring the API can handle large datasets efficiently.

---

## 🧪 Testing Search

You can test the search feature using either `curl` or **Postman**.

### Using curl
When using `curl`, remember that spaces in the query must be manually URL-encoded as `+` or `%20`:
```bash
curl "http://localhost:7000/api/profiles/search?q=young+males+from+nigeria"
```

### Using Postman
Postman is the recommended way to test as it handles URL encoding automatically.
1.  **Method**: `GET`
2.  **URL**: `http://localhost:7000/api/profiles/search`
3.  **Params**: 
    *   Key: `q`
    *   Value: `young males from nigeria` (You can type spaces normally here)
4.  **Pagination**: Add `page` and `limit` keys to the params for paginated results.

## 📄 License
This project is licensed under the MIT License.
