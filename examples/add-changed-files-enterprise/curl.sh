curl -v \
-H 'X-GitHub-Enterprise-Host: github.ford.com' \
-H 'X-GitHub-Event: pull_request' \
-H 'Content-Type: application/json' \
-d "@./pr-payload/payload.json" \
http://localhost:8080

curl -v \
-H 'X-GitHub-Enterprise-Host: github.ford.com' \
-H 'X-GitHub-Event: push' \
-H 'Content-Type: application/json' \
-d "@./push-payload/payload.json" \
http://localhost:8080