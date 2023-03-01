# squid-with-api
A wrapper around Squid Proxy with API exposing some configuration options.

# Missing Tests
1. HTTP PUT verb - expect 400 when path doesn't match body
2. HTTP DELETE verb
3. HTTP PUT all
4. Sync action (HTTP POST) - causes: `squid -k reconfigure`