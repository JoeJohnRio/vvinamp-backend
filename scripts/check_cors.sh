#!/usr/bin/env bash
set -euo pipefail

BASE_URL_LOCAL="http://localhost:8080"
BASE_URL_PUBLIC="https://cirio.my.id"
ORIGIN_HEADER="http://localhost:8081"

echo "=== Local GraphQL preflight ==="
curl -i -X OPTIONS -H "Origin: ${ORIGIN_HEADER}" -H "Access-Control-Request-Method: POST" "${BASE_URL_LOCAL}/v1/graphql" || true

echo "\n=== Local GraphQL POST ==="
curl -i -X POST -H "Origin: ${ORIGIN_HEADER}" -H "Content-Type: application/json" -d '{"query":"{ __typename }"}' "${BASE_URL_LOCAL}/v1/graphql" || true

echo "\n=== Local Storage HEAD (example) ==="
curl -i -I -H "Origin: ${ORIGIN_HEADER}" "${BASE_URL_LOCAL}/storage/your/path/to/image.jpg" || true

echo "\n=== Local Artist media GET (example) ==="
curl -i -H "Origin: ${ORIGIN_HEADER}" "${BASE_URL_LOCAL}/artist/<artistId>/cover.jpg" || true

echo "\n=== Local HLS playlist ==="
curl -i -H "Origin: ${ORIGIN_HEADER}" "${BASE_URL_LOCAL}/hls/<trackId>/index.m3u8" || true

echo "\n=== Public GraphQL preflight ==="
curl -i -X OPTIONS -H "Origin: ${ORIGIN_HEADER}" -H "Access-Control-Request-Method: POST" "${BASE_URL_PUBLIC}/v1/graphql" || true

echo "\n=== Public HLS playlist ==="
curl -i -H "Origin: ${ORIGIN_HEADER}" "${BASE_URL_PUBLIC}/hls/<trackId>/index.m3u8" || true

echo "\n=== Done ==="

echo "Note: replace placeholders in script (artistId, trackId, storage path) before running."
#!/usr/bin/env bash
set -eu

HOST_LOCAL="http://localhost:8080"
HOST_PUBLIC="https://cirio.my.id"
ORIGIN="http://localhost:8081"

echo "== OPTIONS /v1/graphql (local)"
curl -i -X OPTIONS "$HOST_LOCAL/v1/graphql" -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: POST" || true

echo "\n== POST /v1/graphql (local, introspection)"
curl -i -X POST "$HOST_LOCAL/v1/graphql" -H "Origin: $ORIGIN" -H "Content-Type: application/json" -d '{"query":"{ __schema { types { name } } }"}' || true

echo "\n== OPTIONS /storage image (local)"
curl -i -X OPTIONS "$HOST_LOCAL/storage/Mobile/Playlist/6e6ea43f-af03-4b83-9119-308b0abc2a95/playlist.jpg" -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: GET" || true

echo "\n== GET /storage image (local)"
curl -i -X GET "$HOST_LOCAL/storage/Mobile/Playlist/6e6ea43f-af03-4b83-9119-308b0abc2a95/playlist.jpg" -H "Origin: $ORIGIN" || true

echo "\n== OPTIONS /album image (local)"
curl -i -X OPTIONS "$HOST_LOCAL/album/38b598b1-e562-4e40-b3bc-fda988a633b1/semoga_hanya_di_mimpi.jpg" -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: GET" || true

echo "\n== GET /album image (local)"
curl -i -X GET "$HOST_LOCAL/album/38b598b1-e562-4e40-b3bc-fda988a633b1/semoga_hanya_di_mimpi.jpg" -H "Origin: $ORIGIN" || true

echo "\n== OPTIONS /hls playlist (local)"
curl -i -X OPTIONS "$HOST_LOCAL/hls/playlist.m3u8" -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: GET" || true

echo "\n== GET /hls playlist (local)"
curl -i -X GET "$HOST_LOCAL/hls/playlist.m3u8" -H "Origin: $ORIGIN" || true

echo "\n== Range request (local)"
curl -i -X GET "$HOST_LOCAL/track/00000000-0000-0000-0000-000000000000/sample.mp3" -H "Origin: $ORIGIN" -H "Range: bytes=0-1023" || true

# Public tests (uncomment if you want to run against public host)
# echo "\n== OPTIONS /album image (public)"
# curl -i -X OPTIONS "$HOST_PUBLIC/album/38b598b1-e562-4e40-b3bc-fda988a633b1/semoga_hanya_di_mimpi.jpg" -H "Origin: $ORIGIN" -H "Access-Control-Request-Method: GET" || true

# echo "\n== GET /album image (public)"
# curl -i -X GET "$HOST_PUBLIC/album/38b598b1-e562-4e40-b3bc-fda988a633b1/semoga_hanya_di_mimpi.jpg" -H "Origin: $ORIGIN" || true
