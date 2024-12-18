# Bucket Operator

Bucket operator is a follow up on my article about Object Storage on Kubernetes Clusters

* [Object Storage on your k8s cluster using S3 API and SeaweedFS](https://medium.com/@jaferson123/object-storage-on-your-k8s-cluster-using-s3-api-and-seaweedfs-8aba0b34f520)

## Description

The Bucket Operator providers a managed away to use S3 Buckets on kubernetes clusters, allowing user to create buckets via Kubernetes Custom Resource and interact with them via the S3 API.

By deploying the Bucket CR a bucket is created wit the name of the CR

```yaml
apiVersion: objectstorage.mystorage.sh/v1alpha1
kind: Bucket
metadata:
  name: bucket-sample4
spec:
  name: bucket-sample
```

## Getting Started

### Prerequisites
- Access to a Kubernetes v1.11.3+ cluster
- SeaweedFS installed on the cluster
- S3 CLI

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/bucket-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**
// TODO How to install
```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/bucket-operator:tag
```



**Create instances of your solution**
// TODO example of the bucket CRD
```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```


## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

