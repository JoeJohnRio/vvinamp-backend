
first to migrate

psql -U vvinampadmin -d vvinamp

\l                  -- list all databases
\c your_database    -- connect to your database
\dt                 -- list tables
\d users            -- describe the 'users' table
SELECT * FROM users LIMIT 10;  -- view the first 10 rows

ngrok http http://localhost:8080

migrate -path=internal/pkg/db/migrations/mysql -database "postgres://vvinampadmin:password@localhost:5432/vvinamp?sslmode=disable" up
