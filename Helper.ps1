export KO_DOCKER_REPO=registry.ford.com/kcloutie/tests
ko apply --sbom=none --base-import-paths -f config
# ko delete -f config


kubectl get deployments

kubectl patch serviceaccount default -p "{`"imagePullSecrets`": [{`"name`": `"kcloutie-pull-secret`"}]}"

$PodName = (kubectl get pods -o name) -replace "pod/", ""
kubectl logs $PodName 

kubectl port-forward service/el-add-changed-files-listener 8080

Push-Location "./examples/add-changed-files"
kubectl apply -f ./add-changed-files-eventlistener-interceptor.yaml
# kubectl apply -f ./examples/add-changed-files-pr/secret.yaml
kubectl apply -f ../rbac.yaml


kubectl get deployments
kubectl get pods
kubectl get events --sort-by='.metadata.creationTimestamp' 


$Headers = @{
  "X-GitHub-Event" = "pull_request"
  # "X-GitHub-Event" = "push"
  "X-GitHub-Enterprise-Host" = "github.ford.com"
}
$Body = '{"action": "opened","number": 1,"pull_request": {"head": {"sha": "28911bbb5a3e2ea034daf1f6be0a822d50e31e73"}},"repository": {"full_name": "IaC/tekton-helper-operator","clone_url": "https://github.com/IaC/tekton-helper-operator.git"}}'
# $Body = '{"repository": {"full_name": "IaC/tekton-helper-operator","clone_url": "https://github.ford.com/IaC/tekton-helper-operator.git"},"commits": [{"added": [],"removed": [],"modified": ["api/v1beta1/tektonhelperconfig_types.go","config/crd/bases/tekton-helper.ford.com_tektonhelperconfigs.yaml","config/samples/tektonhelperconfig-oomkillpipeline.yaml","config/samples/tektonhelperconfig-timeout.yaml","controllers/tektonhelperconfig_controller.go","examples/oom-pipeline.yaml","pkg/github/github.go"]}]}'

Invoke-RestMethod -Uri "http://localhost:8080" -Headers $Headers -Body $Body -ContentType "application/json"


kubectl logs deploy/el-add-changed-files-listener
# $PodName = (kubectl get taskrun -o yaml | ConvertFrom-Yaml).items[0].status.podName
foreach ($TaskRun in (kubectl get taskrun -o yaml | ConvertFrom-Yaml).items) { 
  Write-Output "======================================================================================="
  Write-Output $TaskRun.status.podName
  Write-Output "======================================================================================="
  kubectl logs $TaskRun.status.podName -c step-display
}
kubectl logs $PodName -c step-display

kubectl get taskrun add-changed-files-run-ntt8m -o yaml

foreach ($TaskRun in (kubectl get taskrun -o yaml | ConvertFrom-Yaml).items) { 
  kubectl delete taskrun $TaskRun.metadata.name
}



# kubectl port-forward service/el-add-changed-files-listener 8080

kubectl get EventListener

kubectl scale deployment add-changed-files-interceptor --replicas=0
kubectl scale deployment add-changed-files-interceptor --replicas=1

kubectl scale deployment el-add-changed-files-listener --replicas=0
kubectl scale deployment el-add-changed-files-listener --replicas=1


kubectl get sa tekton-triggers-core-interceptors -n tekton-pipelines -o yaml
kubectl get deployments -n tekton-pipelines
kubectl get pods -n tekton-pipelines
kubectl get events --sort-by='.metadata.creationTimestamp' -n tekton-pipelines


kubectl logs deploy/tekton-triggers-core-interceptors  -n tekton-pipelines
kubectl logs deploy/tekton-pipelines-webhook  -n tekton-pipelines


for POD in $(kubectl get taskruns -o=jsonpath='{.items[*].status.podName}' | grep add-changed-files-)
do
  echo "=========================================================================================="
  echo $POD
  echo "=========================================================================================="
  kubectl logs $POD
done