curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { listRecentHistory(limit: 10) { id contentId contentType contentName contentImage } }"
  }'