## Run sample knative quarkus benchmarks

The source code of the benchmarks are [here](https://github.com/IBM/knative-quarkus-bench).

#### Environment variables of JVM options for benchmark service
If the base image of the benchmark is `openjdk-11`, please use `JAVA_OPTS_APPEND`
Also, you need to add `-XX:-UseParallelGC` with ecah value of gcType when you configure GC, since `-XX:-UseParallelGC` is set in default by this base image.
If it is ubi-minimal, please use `JAVA_OPTIONS`.

#### Benchmark spec for cpe yaml

- configMapName: the name of configmap that the benchmark service refers
- driverImage: the image of a driver pod
- revName: the revision name of the benchmark service to run. This is mainly used to specify a pod of the target knative benchmark to collect its log.
- command: a set of commands that the driver pod executes. You need to include sleep command to wait until the benchmark service pod complete its process and get its log after the process completion.
- defaultOpts: the default JVM options that you want to use in common

The current parameters of JVM options that you can use are follows:
- gcType
- maxHeapSize
- minHeapSize
- gcThreads
- escapeAnalysis

All the parameters are optional.

Here is an example:

```
apiVersion: cpe.cogadvisor.io/v1
kind: Benchmark
metadata:
  name: graph-bfs 
  namespace: default
spec:
  benchmarkOperator:
    name: jvm 
    namespace: default
  benchmarkSpec: |
    configMapName: "bfs-java-env"
    driverImage: driver:latest 
    defaultOpts: "-Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager -XshowSettings:vm -Xlog:gc*:stderr -XX:+HeapDumpOnOutOfMemoryError"
    revName: "graph-bfs-00001" 
    command: "curl -v http://broker-ingress.knative-eventing.svc.cluster.local/default/default -X POST -H \'Ce-Id: test\' -H \'Ce-source: curl\' -H \'Ce-Specversion: 1.0\' -H \'Ce-Type: graph-bfs\' -H \'Content-Type: application/json\' -d \'{\"size\": \"small\"}\'; sleep 20" 
  iterationSpec:
    iterations:
      - name: xmx
        location: ".maxHeapSize"
        values:
        - "-Xmx512m"
        - "-Xmx1G"
      - name: gc
        location: ".gcType"
        values:
        - "-XX:-UseParallelGC -XX:+UseSerialGC"
        - "-XX:-UseParallelGC -XX:+UseG1GC"
        - "-XX:-UseParallelGC -XX:+UseParallelGC"
    sequential: true
```
