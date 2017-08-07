# OpenEBS Kubernetes PV provisioner

## About OpenEBS 

OpenEBS is containerized storage for containers. More details on OpenEBS can be found on [OpenEBS github page](https://github.com/openebs/openebs)

## How to use OpenEBS kubernetes provisioner

### Building OpenEBS provisioner

```
$ make openebs
```

### Create a docker image on local

```
$ make push-openebs-provisoner
```

### Push OpenEBS provisioner image to docker hub

To push docker image to docker hub you need to have docker hub login credentials. You can pass docker credentials and image name as a environment variable.

```
$ export DIMAGE="docker-username/imagename"
$ export DNAME="docker-username"
$ export DPASS="docker-hub-password"
$ make deploy-openebs-provisioner
```
