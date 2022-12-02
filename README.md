# Add Add Changed Files Cluster Interceptor

This folder contains an implementaion of the add-changed-files cluster interceptor that enriches the payload of an incoming request with
the JSON representation of a pull request as returned by the GitHub API.

This implementation uses the ClusterInterceptor interface. It adds the Add Changed Files under the
`extensions.add-changed-files` field.
