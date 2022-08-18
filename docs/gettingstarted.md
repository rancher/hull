# Getting Started

## Simple Installation

### In Rancher (via Apps & Marketplace)

1. Navigate to `Apps & Marketplace -> Repositories` in your target downstream cluster and create a Repository that points to a `Git repository containing Helm chart or cluster template definitions` where the `Git Repo URL` is `https://github.com/rancher/hull` and the `Git Branch` is `main`
2. Navigate to `Apps & Marketplace -> Charts`; you should see two charts under the new Repository you created: `Hull` and `Hull Example Chart`. 
3. Install `Hull` first
4. Install `Hull Example Chart`

### In a normal Kubernetes cluster (via running Helm 3 locally)

1. Install `hull` onto your cluster via Helm to install the Hull Operator

```
helm install -n cattle-helm-system hull charts/hull
```

2. Install `hull-example` to check out a simple Helm chart containing a ConfigMap and a HelmRelease CR that targets the release itself and keeps it locked into place

```bash
helm install -n cattle-helm-system hull-example charts/hull-example
```

### Checking if the HelmRelease works

1. Ensure that the logs of `hull` in the `cattle-helm-system` namespace show that the controller was able to acquire a lock and has started in that namespace
2. Try to delete or modify the ConfigMaps deployed by the `hull-example` chart (`cattle-helm-system/my-config-map` and `cattle-helm-system/my-config-map-2`); any changes should automatically be overwritten and a log will show up in the Hull logs that showed which ConfigMap it detected a change in
3. Run `kubectl describe helmreleases -n cattle-helm-system hull-example`; you should be able to see events that have been triggered on changes.
4. Upgrade the `hull-example` values to change the contents of the ConfigMap; you should see the modifications show up in the ConfigMap deployed in the cluster as well as events that have been triggered on Hull noticing that change (i.e. you should see a `Transitioning` event that is emitted).

## Uninstalling Hull

After deleting the Helm Charts, you may want to manually uninstall the CRDs from the cluster to clean them up:

```bash
kubectl delete crds helmreleases.helm.cattle.io
```

> Note: Why aren't we packaging Hull CRDs in a CRD chart? Since Hull CRDs can be used for other projects (e.g. [rancher/helm-project-operator](https://github.com/rancher/helm-project-operator), [rancher/prometheus-federator](https://github.com/rancher/prometheus-federator), etc.) and Hull itself can be deployed multiple times to the same cluster, the ownership model of having a single CRD chart that manages installing, upgrading, and uninstalling Hull CRDs isn't a good model for managing CRDs. Instead, it's left as an explicit action that the user should take in order to delete the Hull CRDs from the cluster with caution that it could affect other deployments reliant on those CRDs.