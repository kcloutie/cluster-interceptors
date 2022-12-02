## GitHub EventListener

Creates an EventListener that listens for GitHub webhook events.

### Try it out locally:

1. To create the GitHub trigger and all related resources, run:

   ```bash
   kubectl apply -f .
   kubectl apply -f ../rbac.yaml
   ```

1. Port forward:

   ```bash
   kubectl port-forward service/el-add-changed-files-listener 8080
   ```

1. Test by sending the sample payload.

   ```bash
    curl -v \
    -H 'X-GitHub-Event: pull_request' \
    -H 'X-Hub-Signature: sha1=33035a3a8b7b395139881c2654b59cd1e50ab770' \
    -H 'Content-Type: application/json' \
    -d '{"action": "opened","number": 1,"pull_request": {"head": {"sha": "28911bbb5a3e2ea034daf1f6be0a822d50e31e73"}},"repository": {"full_name": "IaC/tekton-helper-operator","clone_url": "https://github.com/IaC/tekton-helper-operator.git"}}' \
    http://localhost:8080
   ```

   The response status code should be `202 Accepted`

   [`HMAC`](https://www.freeformatter.com/hmac-generator.html) tool used to create X-Hub-Signature.

   In [`HMAC`](https://www.freeformatter.com/hmac-generator.html) `string` is the *body payload ex:* `{"action": "opened", "pull_request":{"head":{"sha": "28911bbb5a3e2ea034daf1f6be0a822d50e31e73"}},"repository":{"clone_url": "https://github.com/tektoncd/triggers.git"}}`
   and `secretKey` is the *given secretToken ex:* `1234567`.

1. You should see a new TaskRun that got created:

   ```bash
   kubectl get taskruns | grep github-run-
   ```
