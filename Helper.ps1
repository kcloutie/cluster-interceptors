export KO_DOCKER_REPO=registry.ford.com/kcloutie/tests
ko apply --sbom=none --base-import-paths -f config




kubectl scale deployment add-changed-files-interceptor --replicas=0
kubectl scale deployment add-changed-files-interceptor --replicas=1
