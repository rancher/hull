# Roadmap

The following types of tests are in scope for the expected use cases that Hull can **eventually** address around Helm chart testing, but are not currently supported:

- **Default Checks for Security / Best Practices:** Add support for additional helper functions for chart developers to use to verify that all workloads in a given rendered template manifest do not contain any known CVEs associated with them by running a `trivy` image scan on all images. Also add additional helper functions to encode best practices for workloads in Helm charts.

- **Integration / Vanilla K8s Testing:** On being provided the `KUBECONFIG` of a cluster that already has the Helm chart installed with some `values.yaml` configuration, allow users to run checks on the live cluster that mimic and test user workflows
  - e.g. Mimic a user creating a certain resource via kubectl and validate that some configuration is modified correctly (a Kubernetes resource is created / modified, a log is emitted, a file is created / modified within a container, the output of a specific HTTP call returns the desired result, etc.)