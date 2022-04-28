# Confluent Cloud Operator

This operator is an abstraction of the confluent-cli for cloud, the benefits of this operator is create topics inside a cluster Kafka managed by Conflunt on the cloud.

For run this operator on your kubernetes cluster is necessary follow some steps, one of the is generates the image on your registry with your configuration to access the confluent cloud.

Now let's go to apply this operator, to facilitate the creation of your topics.

### Step By Step

Obviously the primordia step is check out this repo in your machine for execute the commnads.

1. Go for the path the operator, and run the command.

Always you make some update on the files ***_types.go** is an obligation run the instruction.

``` bash
$ make generate

```

> The above makefile target will invoke the controller-gen utility to update the api/v1alpha1/zz_generated.deepcopy.go file to ensure our API’s Go type definitions implement the runtime.Object interface that all Kind types must implement.


2. The next command is to build the manifests of the operator, its command is important because if do some update in the type, problably you change the CRD. Then to reflect the updates run the instruction.

``` bash
$ make manifests

```

3. let's go to enter in another category of commands, in my machine I run the PODMAN containers, then I use the command PODMAN to build images, if you use Docker only needs to replace PODMAN to DOCKER.

``` bash
$ podman build -t controller --build-arg CCLOUD_EMAIL=put-the-email --build-arg CCLOUD_PASSWORD=put-the-password .

```
> Attention we have to args to build the image, _put-the-email_ and _put-the-password_ this two arguments is your login on the cloud cnfluent.

4. Now we have a locally image, then is necessary to tag the image for your registry, in my case is dockerhub.io

``` bash
$ podman tag localhost/controller:latest docker.io/lhsribas/controller:latest

```
5. before to push the image to remote registry repo, is necessary to perfom the login on the registry

``` bash
$ docker login put-your-registry

```
6. Now push the image to the remote docker image registry.

``` bash
$ podman push docker.io/lhsribas/controller:latest

```

7. The last command is to deploy the operator inside your cluster.

``` bash
$ make deploy

```

## On the Kubernetes

Now you have the operator runnig inside of your kubernetes cluster and to test the Operator execute the follow commands.

I told for you the the command _make manifests_ builds the CRD, the CRD are the rules of the game, then look for the CRD and build the **Kind's** to deploy on your namespaces.

If you not change nothing, the resource are:

```yaml
apiVersion: messages.kubbee.tech/v1alpha1
kind: KafkaCluster
metadata:
  name: kafka-cluster-connection
spec:
  clusterName: kubber1 
  environment: development
  tenant: default
```
> The first is the KafkaCluster, the controller go to the cnfluent cloud and gets the __*ClusterId*__ by the __*ClusterName*__ that you provided, gets the __*EnvironmentId*__ by the name __*Environment*__ that you provided and creates a secret with this information. The secret name always will be __*kafka-cluster-connection*__. The __*tenant*__ is to diferenciate same names for distinct customer and is concated in the topic name with you provides blank, don´t put nothing.

```yaml
apiVersion: messages.kubbee.tech/v1alpha1
kind: KafkaTopic
metadata:
  name: kafkatopic-cadastro-cliente-efetivado
spec:
  # TODO(user): Add fields here
  topicName: cadastro-cliente-efetivado
  partitions: 6
```

> This second is KafkaTopic, the controller gets the secret created by the first kafkacluster controller and use the information to creates the topic inside of your kafka cluster, important: the name of the topic is generated by the tenant-namespace-topicName, afetr the topic creation the controller will be create an configmap with the same name of the KafkaTopic provided on the metadata, and the configmap will contained the topic name and schema registry endpoint.

Now let's go create our namespace and test the operator.

1. Creates the namespace

```bash
$ kubectl create namespace apps
```

2. Apply the KafkaCluster
```bash
$ kubectl apply -f kafkacluster.yml -n apps
```

3. Apply the KafkaTopic
```bash
$ kubectl apply -f kafkatopic.yml -n apps
```

> If you delete the operator, or delete the kind of your project, the operator won´t delete the configmap and secret. And sure not delete your topic at Confluent Cloud Kakfa Cluster.


## Thank you
Thank you, with you have some more questions send me and email
