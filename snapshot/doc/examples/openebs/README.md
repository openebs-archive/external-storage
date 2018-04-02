## Quick Howto

#### Start Snapshot Controller

* Note : Export the Maya-apiserver address as env variable
`export MAPI_ADDR=http://172.18.0.5:5656`

(assuming you have a running Kubernetes local cluster):

```
_output/bin/snapshot-controller  -kubeconfig=${HOME}/.kube/config
```

* Start provisioner (assuming running Kubernetes local cluster):

```bash
 _output/bin/snapshot-provisioner  -kubeconfig=${HOME}/.kube/config
```

Prepare a PV to take snapshot. You can either use OpenEBS dynamic provisioned PVs or static PVs.

```bash
kubectl get pvc
NAME              STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
openebs-vol1-claim   Bound     pvc-f1c1fdf2-00d2-11e8-acdc-54e1ad0c1ccc   5G         RWO            openebs-percona   29s
```

####  Create a snapshot
Now we have PVC bound to a PV that contains some data. We want to take snapshot of this data so we can restore the data later.

* Create a Snapshot for OpenEBS volume Resource

```bash
$ kubectl create -f examples/openebs/snapshot.yaml
```

#### Check VolumeSnapshot and VolumeSnapshotData are created

* Appropriate Kubernetes objects are available and describe the snapshot (output is trimmed for readability):

```yaml
$ kubectl get volumesnapshot,volumesnapshotdata -o yaml
 apiVersion: v1
 items:
   - apiVersion: volumesnapshot.external-storage.k8s.io/v1
  kind: VolumeSnapshot
  metadata:
    clusterName: ""
    creationTimestamp: 2018-01-24T06:58:38Z
    generation: 0
    labels:
      SnapshotMetadata-PVName: pvc-f1c1fdf2-00d2-11e8-acdc-54e1ad0c1ccc
      SnapshotMetadata-Timestamp: "1516777187974315350"
    name: snapshot-demo
    namespace: default
    resourceVersion: "1345"
    selfLink: /apis/volumesnapshot.external-storage.k8s.io/v1/namespaces/default/volumesnapshots/fastfurious
    uid: 014ec851-00d4-11e8-acdc-54e1ad0c1ccc
  spec:
    persistentVolumeClaimName: demo-vol1-claim
    snapshotDataName: k8s-volume-snapshot-2a788036-00d4-11e8-9aa2-54e1ad0c1ccc
  status:
    conditions:
    - lastTransitionTime: 2018-01-24T06:59:48Z
      message: Snapshot created successfully
      reason: ""
      status: "True"
      type: Ready
    creationTimestamp: null
- apiVersion: volumesnapshot.external-storage.k8s.io/v1
  kind: VolumeSnapshotData
  metadata:
    clusterName: ""
    creationTimestamp: 2018-01-24T06:59:48Z
    name: k8s-volume-snapshot-2a788036-00d4-11e8-9aa2-54e1ad0c1ccc
    namespace: ""
    resourceVersion: "1344"
    selfLink: /apis/volumesnapshot.external-storage.k8s.io/v1/k8s-volume-snapshot-2a788036-00d4-11e8-9aa2-54e1ad0c1ccc
    uid: 2a789f5a-00d4-11e8-acdc-54e1ad0c1ccc
  spec:
    openebsVolume:
      snapshotId: pvc-f1c1fdf2-00d2-11e8-acdc-54e1ad0c1ccc_1516777187978793304
    persistentVolumeRef:
      kind: PersistentVolume
      name: pvc-f1c1fdf2-00d2-11e8-acdc-54e1ad0c1ccc
    volumeSnapshotRef:
      kind: VolumeSnapshot
      name: default/snapshot-demo
  status:
    conditions:
    - lastTransitionTime: null
      message: Snapshot created successfully
      reason: ""
      status: "True"
      type: Ready
    creationTimestamp: null
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

* The snapshot is available on host in `/var/openebs/pvc-f1c1fdf2-00d2-11e8-acdc-54e1ad0c1ccc` directory:

## Snapshot based PV Provisioner

Unlike exiting PV provisioners that provision blank volume, Snapshot based PV provisioners create volumes based on existing snapshots. Thus new provisioners are needed.

There is a special annotation give to PVCs that request snapshot based PVs. As illustrated `snapshot.alpha.kubernetes.io` must point to an existing VolumeSnapshot Object
```yaml
metadata:
  name:
  namespace:
  annotations:
    snapshot.alpha.kubernetes.io/snapshot: snapshot-demo
```
### Start PV Provisioner and Storage Class to restore a snapshot to a PV

* Start provisioner (assuming running Kubernetes local cluster):
    ```bash
    _output/bin/snapshot-provisioner  -kubeconfig=${HOME}/.kube/config
    ```

* Create a storage class:
    ```bash
    kubectl create -f examples/openebs/snapshot_sc.yaml
    ```

# Restore a snapshot to a new PV

* Create a PVC that claims a PV based on an existing snapshot
    ```bash
    kubectl create -f examples/openebs/snapshot_claim.yaml
    ```
* Check that a PV was created

    ```bash
    kubectl get pv,pvc
    ```

Snapshots are restored to `/var/openebs/pvc-<name>`.

### Delete the Snapshot:

```bash
$ kubectl delete volumesnapshot/snapshot-demo
volumesnapshot "snapshot-demo" deleted

```
Verify the volumesnapshot object:

```bash
$ kubectl get volumesnapshot -o yaml
apiVersion: v1
items: []
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

