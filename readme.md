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

//run seeder
go run cmd/seeder/main.go reset-and-seed-music

id for user joel@joel.joel
5d7f9219-26cc-4e77-8e1a-fcc38fbea99a

//run the admin-panel
npm run dev