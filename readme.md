
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

ngrok http http://localhost:8080
go run github.com/99designs/gqlgen generate
