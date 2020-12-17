This repository is home for Kubenretes external-storage based Volume provisioners. 

The early engines Jiva and cStor were using the provisioners built from this repository. As Kubernetes community has retired this upstream repository, OpenEBS community are in the process of migrating the code from this repo to component specific repositories. 

If you are looking for the OpenEBS K8s Provisioner, it has already been migrated to https://github.com/openebs/openebs-k8s-provisioner.

This repository is only used for building the snapshot-provionsers used by cStor pools provisioned with SPC. 

_Note: The snapshot provisioners are already deprecated by Kubernetes and will soon be deprecated by the OpenEBS community in favor of the cStor CSI Driver available at https://github.com/openebs/cstor-operators_

For further questions or if you need any help, please reach out to the maintainers via [Kubernetes Slack](https://kubernetes.slack.com).
  * Head to our user discussions at [#openebs](https://kubernetes.slack.com/messages/openebs/)
  * Head to our contributor discussions at [#openebs-dev](https://kubernetes.slack.com/messages/openebs-dev/)

