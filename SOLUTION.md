# Insighta Labs+ System Optimization & Data Ingestion

## 1. Query Performance Optimization

### Approach & Design Decisions
To achieve sub-500ms responses and handle heavy query traffic, i focused on preventing the database from doing redundant or heavy lifting:

1. **Connection Pooling Tuning**: 
   - **Trade-off/Decision**: By default, `pgxpool` limits max connections to 10. Under a load of 50 concurrent requests, this caused a bottleneck where queries waited in line (reaching up to 5.4s latency). We increased `MaxConns` to 100 to handle higher parallel throughput without exhausting Postgres memory.

2. **Database Indexing**:
   - **Trade-off/Decision**: Searching a 1M+ row table caused slow Sequential Scans. We added individual indexes on high-cardinality filtering columns (`country_id`, `gender`, `age`) and a composite index to rapidly serve complex filters without scanning the entire disk. Indexes increase write times and storage, but since read traffic dominates, the massive read-speed benefit outweighs the ingestion cost.

3. **Redis Caching (Upstash)**: 
   - **Trade-off/Decision**: Even with indexes, 1000 concurrent users asking for `country=Nigeria` means 1000 identical database queries. We implemented a cache-first strategy using Redis (Upstash). The first request queries PostgreSQL and stores the result in Redis with a 5-minute TTL. All subsequent identical requests are served directly from Redis without touching the database. The trade-off is a small risk of stale data (up to 5 minutes), but since read traffic dominates and data ingestion is batch-based, this is acceptable.

### Performance Comparison (Query: `?country=Nigeria`)
*Load Test: 1000 requests, 50 concurrent (`hey -n 1000 -c 50`)*

| Metric | Before Optimization (Unoptimized) | After Indexing & Pooling | After Redis Caching |
| :--- | :--- | :--- | :--- |
| **Total Execution Time** | 47.25 secs | 15.98 secs | 11.59 secs |
| **Average Latency** | 2.33 secs | 0.73 secs | 0.55 secs |
| **Slowest Request** | 5.44 secs | 5.48 secs | 1.09 secs |
| **Requests / Second** | 21.15 req/sec | 62.54 req/sec | 86.22 req/sec |

---

## 2. Query Normalization & Cache Efficiency
### Approach & Design Decisions
Users can search using varied natural language phrases (e.g., "Nigerian females" vs "Women in Nigeria"). Caching the raw search string would result in cache misses for identical intents.

To solve this, we parse all natural language queries into a canonical struct (`db.ListProfilesAdvancedParams`). Before executing a database query, we serialize this struct into JSON and generate an MD5 hash. This hash acts as the Redis cache key.
- **Trade-off/Decision**: Using MD5 adds a tiny bit of CPU overhead during hashing, but it guarantees a mathematically identical cache key for any phrases that result in the same filters, drastically increasing Cache Hit rates and ensuring deterministic cache behavior without using AI on every request.

## 3. CSV Data Ingestion
### Approach & Design Decisions
To ingest up to 500,000 rows quickly without degrading the system, we implemented a background-friendly batch streaming approach.

- **Streaming (`encoding/csv`)**: The file is read line-by-line off the incoming HTTP network stream. This prevents the server from loading a massive file into memory, keeping RAM usage low.
- **Chunked Processing (`pgx.Batch`)**: Inserting rows one-by-one is computationally expensive. We group valid rows into chunks of 1000. Once a chunk is full, a single `pgx.Batch` command is sent to Postgres to insert them all concurrently.
- **Handling Failures & Edge Cases**:
  - We validate every row as it is read. Missing names, negative ages, or invalid gender strings simply increment the `skipped` counter and are bypassed.
  - To handle duplicate names without failing the entire 1000-row batch (partial failure requirement), we use PostgreSQL's `ON CONFLICT (name) DO NOTHING`. This gracefully ignores existing names while inserting the rest of the batch perfectly.
  - A comprehensive JSON summary is tracked in-memory and returned upon completion, reporting exactly what was successfully inserted and why rows were rejected.
