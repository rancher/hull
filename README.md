hull
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Hull is a **Go testing framework** for writing comprehensive tests on [Helm](https://github.com/helm/helm) charts.

Once you have defined your suite of tests targeting a specific chart (or multiple charts) using Hull, **you can simply run your suite(s) of tests by running `go test`**.

## Who needs Hull?

Anyone who maintains a Helm repository or a set of Helm charts that would like to add an automated testing suite that allows them to **lint** charts and setup **unit tests**.

For more information on why you might want to use Hull, see the [About guide](docs/about.md).

## Prerequisites

You will be expected to install the following dependencies locally on your machine to successfully run Hull:
* [Go](https://go.dev) (minimal requirement to be able to run `go test`)
* [Yamllint](https://github.com/adrienverge/yamllint) (only required if you use Hull to run YAML linting on manifests produced by `helm template` commands)

## Getting Started

Please see [`examples/example_test.go`](./examples/example_test.go) for an example of a Go test written for a single chart in this fashion on the chart located in [`testdata/charts/example-chart`](./testdata/charts/example-chart/). To run the example test, you can simply run:

```bash
go test examples/example_test.go
```

Under the hood, Hull leverages [`github.com/stretchr/testify/assert`](github.com/stretchr/testify/assert) for test assertions; it is recommended, but not required, for users to also use this framework when designing Hull tests.

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
