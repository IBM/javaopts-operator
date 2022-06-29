# Javaopts-operator
## Operator for benchmark testing of Java-based Knative applications
This operator is designed for running [knative-quarkus-bench](https://github.com/IBM/knative-quarkus-bench) with [CPE](https://github.com/IBM/cpe-operator).
You can also run this operator w/o CPE.

#### What this operator do
- Create/Update a configMap of JVM options that a Knative app service refers
- Delete the pods of the target Knative app to apply the configMap
- Create and Run a driver job for your Knative app that sends requests and collect logs of the target Knative app pod

#### How to use this operator as a benchmark operator of CPE
0. Run Knative service of your application
```
$ kubectl create -f broker.yaml -f sleep_ksvc.yaml -f sleep_trigger.yaml
```
If the process completion of your knative application after traffic ends would take longer than 30 secs, 
you had better to disable `enable-scale-to-zero` or increase `scale-to-zero-grace-period` in your environment.
For each service, you can configure `scale-to-zero-pod-retention-period` as below:
```
kind: Service
apiVersion: serving.knative.dev/v1
metadata:
  name: sleep 
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/scale-to-zero-pod-retention-period: "5m"
    spec:
      containers:
......
```

1. Deploy [CPE](https://github.com/IBM/cpe-operator)

2. Build an image of this operator if needed
   Please see [Makefile](Makefile) to configure image registry
```
$ cd javaopts-operator
$ make generate
$ make manifests
$ make docker-build docker-push
```
3. Run this operator
```
$ kubectl create -f jvmopts_operator.yaml
```
You can found their resources in the namespace `jvmopts-operator`. 
```
$ kubectl get all -n jvmopts-operator
NAME                                    READY   STATUS    RESTARTS   AGE
pod/jvmopts-operator-6f549bcbdd-b4gsm   1/1     Running   0          43s

NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/jvmopts-operator   1/1     1            1           43s

NAME                                          DESIRED   CURRENT   READY   AGE
replicaset.apps/jvmopts-operator-6f549bcbdd   1         1         1       43s
```

4. Create benchmark resouce for CPE
```
$ kubectl create -f cpe_knative_sleep.yaml
```
Then, CPE creates jobs of a driver for your application.


#### How to use this operator w/o CPE
0. Run Knative service of your application
1. Build an image of this operator if needed
   Please see [Makefile](Makefile)
```
$ cd javaopts-operator
$ make generate
$ make manifests
```
2. Deploy 
```
$ make deploy
```
