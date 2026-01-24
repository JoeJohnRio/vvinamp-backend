# Music Data Seeder

This seeder populates the database with Indonesian music data including artists, albums, tracks, and listening history from popular Indonesian artists.

## Artists Included

- **DEWA 19** - Iconic Indonesian rock band
- **Sheila On 7** - Popular Indonesian pop rock band
- **Bernadya** - Contemporary Indonesian artist
- **Hindia** - Indie pop/rock artist

## Features

- **Replace all music data**: Clears existing music data while preserving user data
- **Realistic data**: Includes actual albums, songs, and lyrics from these artists
- **Listening history**: Seeds listening history for the admin user with ID `5d7f9219-26cc-4e77-8e1a-fcc38fbea99a`
- **Image integration**: Uses artist photos from your VvinampData folder
- **Proper relationships**: Maintains correct relationships between artists, albums, and tracks

## Usage

To seed music data (replacing existing music data):

```bash
go run cmd/seeder/main.go reset-and-seed-music
```

Alternative command (does the same thing):
```bash
go run cmd/seeder/main.go seed-music
```

Both commands will:
- Clear all existing music data (artists, albums, tracks, listening history)
- Preserve all user data
- Seed new Indonesian music data
- Create listening history for the admin user

## Notes

- The seeder uses image files from `S:/Project German/VvinampData/` folder for artist photos
- Album covers use placeholder images from the same folder
- Stream counts and other metrics are realistic but simulated
- Each artist has their actual popular songs included
- Listening history is created for the specified admin user ID
- All music-related tables are cleared except user data