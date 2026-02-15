# Media File Serving API

This Go application implements three HTTP GET endpoints using the Gin framework to serve media files based on artist, album, and track information stored in a PostgreSQL database.

## Endpoints

### 1. Artist Media Endpoint
```
GET /artist/:artistId/:fileName
```
- Retrieves the artist name by ID from the database
- Constructs path: `<BaseStorage>/artist/<sanitized_artist_name>/<fileName>`
- Example: `/artist/123e4567-e89b-12d3-a456-426614174000/hindia.jpg`

### 2. Album Media Endpoint
```
GET /album/:albumId/:fileName
```
- Retrieves album title and artist name by album ID from the database
- Constructs path: `<BaseStorage>/artist/<sanitized_artist_name>/album/<sanitized_album_title>/<fileName>`
- Example: `/album/987fc8a2-1d9c-49e2-8a00-123456789abc/lagipula_hidup_akan_berakhir.jpg`

### 3. Track Media Endpoint
```
GET /track/:trackId/:fileName
```
- Retrieves track title, album title, and artist name by track ID from the database
- Constructs path: `<BaseStorage>/artist/<sanitized_artist_name>/album/<sanitized_album_title>/tracks/<fileName>`
- Example: `/track/456de789-2e0f-410a-b123-fedcba987654/masalah_masa_depan.mp3`

## Sanitization Rules

The application applies the following sanitization rules to artist and album names:
1. Replace any character NOT in [a-zA-Z0-9 space - _] with "_"
2. Convert to lowercase
3. Replace multiple spaces/underscores with a single "_"

Examples:
- "Hindia" → "hindia"
- "Lagipula Hidup akan Berakhir" → "lagipula_hidup_akan_berakhir"

## Configuration

Configuration is handled via environment variables in the `.env` file:

```env
# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/vvinamp?sslmode=disable

# Storage Configuration
BASE_STORAGE_PATH=S:\Project German\Backend\Storage

# Server Configuration
PORT=8080
GIN_MODE=debug
```

## Dependencies

- Go 1.22+
- Gin framework
- PostgreSQL driver (pgx/v5)
- UUID package (google/uuid)
- Environment loader (godotenv)

## Security Features

- Directory traversal prevention using `filepath.Clean()`
- Input validation for UUID parameters
- File existence verification before serving
- Safe path construction using `filepath.Join()`

## Running the Application

1. Ensure PostgreSQL is running and accessible with the configured credentials
2. Install dependencies: `go mod tidy`
3. Run the application: `go run main.go handlers.go db.go`

For testing without a database, use: `go run main_nodb.go handlers.go db.go`