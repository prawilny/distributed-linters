WARNING: This allows any user with read access to secrets or the ability to create a pod to access super-user credentials.
kubectl create clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts

eval $(minikube -p minikube docker-env)

make

``` mm.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: machine-manager
  labels:
    app: irio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: irio
  template:
    metadata:
      labels:
        app: irio
    spec:
      containers:
      - name: machine-manager
        image: machine_manager_container
        imagePullPolicy: Never
        ports:
        - containerPort: 2137
```

kubectl apply -f mm.yaml

kubectl expose deployment machine-manager --type=LoadBalancer --name=machine-manager-exposed

kubectl get services <- IP

minikube tunnel

./admin IP:2137 add_version -image_url=python_linter_container -language=python -version=1.0