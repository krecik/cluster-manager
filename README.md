# kubecare cluster manager

kubecare cluster manager is a tool for generating manifests used by ArgoCD.
By using kubecare cluster manager you'll get the following benefits:

* simplified syntax for defining cluster applications
* you don't repeat yourself with managing multiple clusters
* auto creation of namespaces for helm charts
* addon support to extract common cod
* support for Helm and Kustomize (as well as plain manifests)
* support for multiple variants of addons

## Usage

After installing and configuring ArgoCD to work with kubecare cluster manager you need to:

### Create new git repository

You can use any provider that ArgoCD can connect to.

### Configure ArgoCD to be able to read from that repo

In order to allow ArgoCD to connect to the repo you need to add private ssh key for that repo in ArgoCD configuration as well as add public key in your git repo configuration.
See Installation part below to see how to configure ssh private key in ArgoCD.

### Create basic structure inside the repo

```bash
CLUSTER_NAME=my-cluster
mkdir clusters
mkdir addons
mkdir clusters/$CLUSTER_NAME
mkdir clusters/$CLUSTER_NAME/addons
touch clusters/$CLUSTER_NAME/cluster.yaml
```

### Edit cluster definition file

Now you need to edit _clusters/$CLUSTER_NAME/cluster.yaml_ file and provide some basic information:

```yaml
cluster:
  name: my-cluster
  server: https://url-to-kube-api-server
  # optional: repoURL: url-to-external-repo, by default it's the same url as the repo with cluster.yaml file

helmApplications:
# see below
kustomizeApplications:
# see below
  
```

### Define kustomize application

### Define helm application

### Create helm addon

### Splitting cluster definition file into multiple files


## Installation on ArgoCD

TODO
