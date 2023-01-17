# anynines-homework
Basically you will find here how to run and install the dummy operator inside your kubernetes cluster

## Prerequisite
1. Local kubernetes cluster ,  You can use [minikube](https://minikube.sigs.k8s.io/docs/start/)
2. Operator SDK CLI, [Download](https://sdk.operatorframework.io/docs/installation/)
3. Kubectl CLI installed, [Here](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
4. Docker, [Download](https://docs.docker.com/get-docker/)
5. GO, [Download](https://go.dev/dl/)

## Getting Started
After installing the Prerequisites tools, we will need to follow some set of steps to have our controller deployed intp the cluster.
**Note:** We have 2 options to run our kubernetes cluster:
1. Run it outside the cluster, your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).
2. Run inside the cluster as a deployment.


### Running the operator
In a simple words, we will deal with multiple main folders:
1. api/v1alpha1, for editning our types structures for the needed specs or status.
2. config/crd/bases/interview.com_dummies.yaml file to install the CRD inside our cluster.
3. conrollers folder, to modify our reconcile loop mechanism.
4. tobeinstall folder for the needed manifests to be installed inside our cluster to make the operator ready.

### Let's Start
1. Install Instances of Custom Resources::
```sh
kubectl apply -f tobeinstall/5-interview.com_dummies_Crd.yaml
```
**To run the operator outside the cluster :** Just run from the root of the project, or Proceed to step 2.
```sh
go run main.go
```
2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/anynines-homework:tag
```
**Ex:**	make docker-build IMG=mohamednabiel717/controller:0.0.1,make docker-push IMG=mohamednabiel717/controller:0.0.1

3. Deploy the manifiests under tobeinstall/ folder with the order found.

```sh
kubectl apply -f tobeinstall/1,2,3,4,6,7-
```
**Note:** We already installed 5-interview.com_dummies_Crd.yaml above.


### Test It Out 
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```
