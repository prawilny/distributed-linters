WARNING: This allows any user with read access to secrets or the ability to create a pod to access super-user credentials.
kubectl create clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts

eval $(minikube -p minikube docker-env)

minikube tunnel

make

kubectl apply -f yaml/

kubectl get services <- IP

./admin IP:2137 add_version -image_url=python_linter_container -language=python -version=1.0