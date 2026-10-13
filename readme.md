first to migrate

psql -U vvinampadmin -d vvinamp

migrate -path=internal/pkg/db/migrations/mysql -database "postgres://vvinampadmin:password@localhost:5432/vvinamp?sslmode=disable" up

\l                  -- list all databases
\c your_database    -- connect to your database
\dt                 -- list tables
\d users            -- describe the 'users' table
SELECT * FROM users LIMIT 10;  -- view the first 10 rows

gqlgen generate // to GENERATE
go run -mod=mod github.com/99designs/gqlgen generate

//run cloudflare
cloudflared tunnel run cirio-laptop

//run redis
docker run -d --name my-redis -p 6379:6379 redis:latest

//run seeder
go run cmd/seeder/main.go reset-and-seed-music

id for user joel@joel.joel
5d7f9219-26cc-4e77-8e1a-fcc38fbea99a

//run the admin-panel
npm run dev

## CORS configuration (GraphQL)

The API now exposes CORS handling for `/v1/graphql`. You can configure allowed origins and whether credentials are accepted via environment variables:

- `CORS_ALLOW_ORIGINS` — comma-separated list of allowed origins. Default: `http://localhost:8081`
- `CORS_ALLOW_CREDENTIALS` — `true`/`false` (default `true`). When true the server will echo the exact `Origin` header (do not use `*`).

Examples:

Set development origins:

```bash
export CORS_ALLOW_ORIGINS="http://localhost:8081"
export CORS_ALLOW_CREDENTIALS=true
```

Set production + staging:

```bash
export CORS_ALLOW_ORIGINS="https://app.staging.example.com,https://app.example.com"
export CORS_ALLOW_CREDENTIALS=true
```

Quick tests (replace host if needed):

Preflight OPTIONS test:

```bash
curl -i -X OPTIONS https://cirio.my.id/v1/graphql \
	-H "Origin: http://localhost:8081" \
	-H "Access-Control-Request-Method: POST"
```

GraphQL POST test (simple introspection):

```bash
curl -i -X POST https://cirio.my.id/v1/graphql \
	-H "Origin: http://localhost:8081" \
	-H "Content-Type: application/json" \
	-d '{"query":"{ __schema { queryType { name } } }"}'
```

If you are using a reverse proxy (Nginx, Cloudflare Tunnel, Cloudflare Workers, etc.), ensure it does not strip the CORS response headers. For Nginx add `add_header` directives at the `location` that proxies `/v1/graphql` and include `always` so headers are present on error responses as well.