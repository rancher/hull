hull
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Hull is a tool for testing Helm charts.

## Prerequisites

Similar to [`helm/chart-testing`](https://github.com/helm/chart-testing), it is recommended to use the `rancher/hull` Docker image that already comes with the following tools pre-installed. However, if you are running the Hull binary locally, you will be expected to install the following dependencies locally on your machine to successfully run Hull:
* [Go](https://go.dev)
* [Trivy](https://github.com/aquasecurity/trivy)
* [Yamllint](https://github.com/adrienverge/yamllint)

## Getting Started

For more information, see the [Getting Started guide](docs/gettingstarted.md).

## Who needs Hull?

Anyone who maintains a Helm repository that would like to add an automated testing suite that covers the following basic, core functionality:

### Chart Testing (Batteries Included)

- **Linting + Schema Validation:** Ensure that the Helm chart passes a basic `helm lint` command and that basic, common values.yaml configurations of the Helm charts compile to YAML that passes stylistic linting checks, without validating the actual content of the YAML produced from a logical perspective.
- **Unit Testing:** Ensure that the Kubernetes manifest produced by running the `helm template` command on basic, common values.yaml configurations of the Helm charts compiles to YAML that produces valid content that matches the Kubernetes schema.
- **Security Testing:** Ensure that any images provided to the chart in basic, common values.yaml configurations of the Helm charts (including the default values.yaml configuration) do not contain any known CVEs associated with them by running a trivy image scan on them

### Release Testing (Batteries Included)

- **Install Testing / Smoke Testing:** Upon installing a helm chart onto a cluster using basic, common values.yaml configurations of Helm charts, ensure that all pods are up and running (`helm install --wait`) and that a `helm test` is successfully executed. Also ensure that upgrades from the second newest version of a chart (if it exists) to the latest version of the chart work if a given values.yaml exists in both the newer and older chart and that in-place upgrades work.
  - Note: it is expected that your Helm chart must already have `helm test` hook resources that verify that all exposed endpoints of your application can be accessed via some k8s Service using Kubernetes DNS directly from within the cluster and that any known endpoints that indicate application status (i.e. a /health endpoint) return a valid response that indicates that the endpoint is healthy, if available. See the [Grafana Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/grafana/templates/tests) or [MySQL Helm chart](https://github.com/helm/charts/tree/master/stable/mysql/templates/tests) for simple examples of how to add this type of testing via [Bats](https://github.com/bats-core/bats-core)

### Advanced Testing (Batteries NOT Included)

Since Hull just runs all the Go tests mounted to the container at the `test` directory, users are also recommended to implement the following additional application-specific / business-logic-specific suite of tests:

- **Advanced Unit Testing:** Tests that are run on the sets of Kubernetes manifests produced by running the `helm template` command on basic, common values.yaml configurations of the Helm charts
  - e.g. Resources follow best practices that allow them to be deployed onto specialized environments (i.e. Deployments have nodeSelectors and tolerations for Windows)
  - e.g. Resources meet other business-specific requirements (i.e. all images start with rancher/mirrored- unless they belong on an allow-list of Rancher original images)
- **Integration / Vanilla K8s Testing:** Tests that are run upon installing a helm chart onto a cluster using basic, common values.yaml configurations of Helm charts by mimicking common user workflows
  - e.g. Mimic a user creating a certain resource via kubectl and validate that some configuration is modified correctly (a Kubernetes resource is created / modified, a log is emitted, a file is created / modified within a container, the output of a specific HTTP call returns the desired result, etc.)

To support defining these Go tests, Hull has a set of helper libraries that define allow users to quickly parse all templates and execute Rego-style policies on resources that are created from those manifests or execute certain common actions on live clusters.

Once your custom tests have been written, either simply run the Hull image with the additional Go tests mounted somewhere in the `/home/hull/tests` directory or build a new image off the `rancher/hull` base image that already contains your tests. 

*In general, it's recommended to package custom tests in your own image to allow them to be run easily locally or as part of GitHub Action Workflows. See `rancher/rancher-hull` for an example of such an image.*

## How is this different from projects like `helm/chart-testing`?

[`helm/chart-testing`](https://github.com/helm/chart-testing) currently only supports the following basic linting and smoke testing (via running `helm install|upgrade` followed by `helm test`) for Helm charts:
- Linting the `Chart.yaml` fields via [Yamale](https://github.com/23andMe/Yamale)
- Running [Yamllint](https://github.com/adrienverge/yamllint) on the `Chart.yaml` and all `values.yaml` files (including those in `ci/*-values.yaml`)
- Validating that the maintainers listed in the Chart.yaml are valid accounts
- Running `helm lint` on all `values.yaml` files (including those in `ci/*-values.yaml`)
- Running various `helm install` / `helm-upgrade` scenarios on all `values.yaml` files (including those in `ci/*-values.yaml`), where each action is followed by a `helm test` that will run all the test hooks defined in the chart

Hull supports most of these capabilities but makes it easier to add more advanced testing functionality by allowing users to add additional Go tests using a framework of your own choice!

See the [Design guide](docs/design.md) for more information on what kinds of tests are covered by default.

## Developing

### Which branch do I make changes on?

Hull is built and released off the contents of the `main` branch. To make a contribution, open up a PR to the `main` branch.

For more information, see the [Developing guide](docs/developing.md).

## Building

`make`


## Running

`./bin/hull`

## License
Copyright (c) 2022 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
