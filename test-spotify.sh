#!/bin/bash
# Test Spotify API directly from inside the container
set -e

# Get token
echo "=== Getting token ==="
TOKEN=$(curl -s -X POST https://accounts.spotify.com/api/token \
  -d "grant_type=client_credentials" \
  -u "844a34aaf04247b4a3afa094d60ef57a:88d2c617531b46f78b8fbf4edd6d6aae" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
echo "Token: ${TOKEN:0:15}..."

# Search
echo ""
echo "=== Search: I MISS YOU ==="
curl -s "https://api.spotify.com/v1/search?q=I+MISS+YOU&type=track&limit=3" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
tracks=d.get('tracks',{}).get('items',[])
print(f'Results: {len(tracks)}')
for t in tracks[:3]:
    print(f'  - {t[\"name\"]} by {t[\"artists\"][0][\"name\"]}')
"

echo ""
echo "=== Search: blink 182 ==="
curl -s "https://api.spotify.com/v1/search?q=blink+182&type=track&limit=3" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
tracks=d.get('tracks',{}).get('items',[])
print(f'Results: {len(tracks)}')
for t in tracks[:3]:
    print(f'  - {t[\"name\"]} by {t[\"artists\"][0][\"name\"]}')
"
