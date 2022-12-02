# Add Add Changed Files Cluster Interceptor

This folder contains an implementaion of the add-changed-files cluster interceptor that enriches the payload of an incoming request with a list of changed files releated to a PR or push.

This implementation uses the ClusterInterceptor interface. It adds the Add Changed Files under the
`extensions.add-changed-files` field.
