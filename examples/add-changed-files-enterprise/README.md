## GitHub Add Changed Files EventListener

> NOTE: This is for internal testing against github enterprise, please use the `add-changed-files` example instead.

Creates an EventListener that listens for push and pull_request GitHub webhook events. It will add the files that were changed to the extension section of the payload

### Try it out locally:

1. To create the GitHub Add Changed Files trigger and all related resources, run:

> NOTE: for authentication support, the `github-secret` must be updated to a valid github personal access token that has read access to the repository

   ```bash
   kubectl apply -f .
   kubectl apply -f ../rbac.yaml
   ```

1. Port forward:

   ```bash
   kubectl port-forward service/el-add-changed-files-listener 8080
   ```

1. Test by sending the pull request sample payload.

   ```bash
    curl -v \
    -H 'X-GitHub-Enterprise-Host: github.ford.com' \
    -H 'X-GitHub-Event: pull_request' \
    -H 'Content-Type: application/json' \
    -d "@./pr-payload/payload.json" \
    http://localhost:8080
   ```

   The response status code should be `202 Accepted`

1. Test by sending the pull request sample payload.

   ```bash
    curl -v \
    -H 'X-GitHub-Enterprise-Host: github.ford.com' \
    -H 'X-GitHub-Event: push' \
    -H 'Content-Type: application/json' \
    -d "@./push-payload/payload.json" \
    http://localhost:8080
   ```

   The response status code should be `202 Accepted`

1. You should see the files changed in the log output of the tasksRun:

   ```bash
    for POD in $(kubectl get taskruns -o=jsonpath='{.items[*].status.podName}' | grep add-changed-files-)
    do
      echo "=========================================================================================="
      echo $POD
      echo "=========================================================================================="
      kubectl logs $POD -c step-display
    done
   ```
