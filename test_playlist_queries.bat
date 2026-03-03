@echo off
REM Test Playlist Feature - GraphQL Queries and Mutations

set GRAPHQL_ENDPOINT=http://localhost:8080/v1/graphql

echo ============================================
echo Testing Playlist Feature
echo ============================================
echo.

REM First, login to get an access token
echo Step 1: Login to get access token...
echo.

set LOGIN_QUERY={\"query\":\"mutation { login(email: \\\"joel@joel.joel\\\", password: \\\"yourpassword\\\") { session { accessToken refreshToken expiresAt } } }\"}

echo Login Query:
echo %LOGIN_QUERY%
echo.
echo Please run this manually in GraphiQL or update the password and run with curl
echo.

echo ============================================
echo Sample GraphQL Operations (copy to GraphiQL)
echo ============================================

echo.
echo --- 1. Create Playlist ---
echo mutation {
echo   createPlaylist(input: {
echo     title: "My Test Playlist"
echo     description: "Testing playlist creation"
echo     isPublic: true
echo   }) {
echo     id
echo     title
echo     description
echo     isPublic
echo     createdAt
echo     creator {
echo       id
echo       username
echo     }
echo   }
echo }

echo.
echo --- 2. Get My Playlists ---
echo query {
echo   getMyPlaylists(first: 10) {
echo     edges {
echo       cursor
echo       node {
echo         id
echo         title
echo         isPublic
echo         savedCount
echo         totalDuration
echo       }
echo     }
echo     pageInfo {
echo       hasNextPage
echo       hasPreviousPage
echo       startCursor
echo       endCursor
echo     }
echo     totalCount
echo   }
echo }

echo.
echo --- 3. Get Playlist by ID ---
echo query {
echo   getPlaylist(id: "YOUR_PLAYLIST_ID_HERE") {
echo     id
echo     title
echo     description
echo     isPublic
echo     bannerImage
echo     coverImage
echo     savedCount
echo     totalDuration
echo     isSavedByMe
echo     creator {
echo       id
echo       username
echo       profilePicture
echo     }
echo     tracks {
echo       track {
echo         id
echo         title
echo         duration
echo         artistName
echo         albumCoverUrl
echo       }
echo       position
echo       addedAt
echo     }
echo   }
echo }

echo.
echo --- 4. Add Song to Playlist ---
echo mutation {
echo   addSongToPlaylist(input: {
echo     playlistId: "YOUR_PLAYLIST_ID"
echo     trackId: "YOUR_TRACK_ID"
echo     position: 1
echo   }) {
echo     id
echo     title
echo     tracks {
echo       track {
echo         id
echo         title
echo       }
echo       position
echo     }
echo   }
echo }

echo.
echo --- 5. Remove Song from Playlist ---
echo mutation {
echo   removeSongFromPlaylist(input: {
echo     playlistId: "YOUR_PLAYLIST_ID"
echo     trackId: "YOUR_TRACK_ID"
echo   })
echo }

echo.
echo --- 6. Save Playlist ---
echo mutation {
echo   savePlaylist(input: {
echo     playlistId: "YOUR_PLAYLIST_ID"
echo   }) {
echo     id
echo     title
echo     savedCount
echo     isSavedByMe
echo   }
echo }

echo.
echo --- 7. Unsave Playlist ---
echo mutation {
echo   unsavePlaylist(input: {
echo     playlistId: "YOUR_PLAYLIST_ID"
echo   })
echo }

echo.
echo --- 8. Update Playlist ---
echo mutation {
echo   updatePlaylist(id: "YOUR_PLAYLIST_ID", input: {
echo     title: "Updated Playlist Title"
echo     description: "Updated description"
echo     bannerImage: "https://example.com/new-banner.jpg"
echo   }) {
echo     id
echo     title
echo     description
echo     bannerImage
echo     updatedAt
echo   }
echo }

echo.
echo --- 9. Delete Playlist ---
echo mutation {
echo   deletePlaylist(id: "YOUR_PLAYLIST_ID")
echo }

echo.
echo --- 10. Get Playlists By User ---
echo query {
echo   getPlaylistsByUser(userId: "USER_ID_HERE", first: 10) {
echo     edges {
echo       node {
echo         id
echo         title
echo         isPublic
echo       }
echo     }
echo     totalCount
echo   }
echo }

echo.
echo ============================================
echo To test:
echo 1. Start the server: go run main.go
echo 2. Open GraphiQL: http://localhost:8080/v1/graphiql
echo 3. Copy-paste the queries above and replace IDs
echo ============================================

pause
