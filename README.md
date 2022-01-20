## Cluster setup
In order to set up the project execute the following steps:
- log into gcloud
- set project id: `gcloud config set project irio-test-338916`
- set docker environment: `gcloud auth configure-docker europe-central2-docker.pkg.dev`
- make sure you can list images in the registry: `gcloud artifacts docker images list europe-central2-docker.pkg.dev/irio-test-338916/irio-test`
- build container images: `make -Bj`
- push container images to artifact registry: `make image_upload -j`
- make sure you can see the images in the registry: `gcloud artifacts docker images list europe-central2-docker.pkg.dev/irio-test-338916/irio-test`
- deploy the application to kubernetes: `kubectl create -f yaml`
- deploy the monitoring to kubernetes: `kubectl create -f prometheus`

## Tests
In order to run the tests execute the following steps:

- Check that the cluster is ready.
- Set the following environment variables with external IPs of appropriate just created services:
  - LOAD_BALANCER=IP:20000
  - MANAGER=IP:10000
  - PROMETHEUS=IP:9090
  - REGISTRY=europe-central2-docker.pkg.dev/irio-test-338916/irio-test
- build admin client: `cd admin && go build`
- enter test directory: `cd ../test`
- run the tests:
  - simple end to end test of machine manager managing linter deployments: `go run test_manager.go`
  - simple test showing that load balancer is working as expected:
    - prepare the environment: create linter deployments and set proportions:
      - `./admin $MANAGER add_version -image_url=$REGISTRY/python_linter -language=python -version=1.0`
      - `./admin $MANAGER add_version -image_url=$REGISTRY/python_linter -language=python -version=2.0`
      - `./admin $MANAGER set_proportions python 1.0 2 python 2.0 1`
    - wait for the pods to start
    - run the test: `go run test_loadbalancing.go`
    - check that the values reported as routed to versions 1.0 and 2.0 have ratio close to 2.
    - if you want to remove the created linters:
      - set their weights to 0: `./admin $MANAGER set_proportions python 1.0 0 python 2.0 0`
      - delete the deployments:
        - `./admin $MANAGER remove_version -language=python -version=1.0`
        - `./admin $MANAGER remove_version -language=python -version=2.0`

## Witnessing autoscaling
In order to check the autoscaling out execute the following steps:
- Set up cluster (with monitoring!)
- Deploy load generating program: `kubectl create -f load/load.yaml`
- Use `kubectl` to watch pods being created
- Use Prometheus web interface (service "prometheus", port 9090) to view metric `load_lints` to assess the system's performance.
- It's suggested to delete the resources afterwards: `kubectl delete -f load/load.yaml`

## Admin CLI tool usage
Arguments for the `admin` tool:

`admin MANAGER_ADDRESS SUBCOMMAND [ARG]...`

- `admin MANAGER_ADDRESS list_versions`
- `admin MANAGER_ADDRESS add_version -image_url=IMAGE -language=LANGUAGE -version=VERSION`
- `admin MANAGER_ADDRESS remove_version -language=LANGUAGE -version=VERSION`
- `admin MANAGER_ADDRESS set_proportions [LANGUAGE VERSION PROPORTION]...`
