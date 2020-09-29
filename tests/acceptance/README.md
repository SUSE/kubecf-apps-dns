# Acceptance Tests

The Apps DNS acceptance tests work by deploying the Apps DNS pod with the 
configuration pointing to the acceptance test job as the Service Discovery
Controller (SDC). This allows us to fake the SDC responses and assert the DNS
answers obtained from the Apps DNS.

## Running the tests locally

1. Start Minikube with the service clusterIP range `10.43.0.0/16`:

```shell
minikube start --service-cluster-ip-range 10.43.0.0/16
```

2. Build the required images using the Minikube Docker daemon:

```shell
git_root=$(git rev-parse --show-toplevel)

(eval $(minikube -p minikube docker-env); docker buildx build -t acceptance-tests -f "${git_root}/tests/acceptance/image/Dockerfile" "${git_root}/tests/acceptance")

(eval $(minikube -p minikube docker-env); docker buildx build -t kubecf-apps-dns -f "${git_root}/image/Dockerfile" "${git_root}")
```

3. Install the quarks-secret:

```shell
kubectl create --filename "${git_root}/tests/acceptance/deploy/k8s/namespaces.yaml"

helm install qsecret quarks/quarks-secret \
    --wait \
    --namespace qsecret \
    --set "global.monitoredID=qsecret"
```

4. Deploy the Apps DNS:

```shell
kubectl apply --filename "${git_root}/tests/acceptance/deploy/k8s/mtls_certificates.yaml"
kubectl apply --filename "${git_root}/tests/acceptance/deploy/k8s/apps_dns.yaml"
```

5. Once the Apps DNS is ready, deploy the tests and tail the logs:

```shell
kubectl apply --filename "${git_root}/tests/acceptance/deploy/k8s/test.yaml"
kubectl logs \
    --follow \
    --namespace tests \
    --selector app=acceptance-tests \
    --tail -1
```
