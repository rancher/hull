# Migrating from Hull to helm-unittest + chart-testing + kubeconform

This guide walks through replacing [Hull](https://github.com/rancher/hull) with a combination of three widely-adopted tools:

| Tool | Replaces |
|------|----------|
| [helm-unittest](https://github.com/helm-unittest/helm-unittest) | Unit tests (test cases, named checks, failure cases) |
| [chart-testing (`ct`)](https://github.com/helm/chart-testing) | Helm lint + CI chart validation |
| [kubeconform](https://github.com/yannh/kubeconform) | Kubernetes manifest schema validation |

**What you lose:** Hull's coverage tracker (which fields in `values.yaml` are exercised by tests) has no direct equivalent in these tools. If coverage tracking is critical, keep Hull for that purpose or track it manually.

---

## Table of Contents

1. [Install the tools](#1-install-the-tools)
2. [Concept mapping](#2-concept-mapping)
3. [Converting test cases](#3-converting-test-cases)
4. [Converting named checks](#4-converting-named-checks)
5. [Converting failure cases](#5-converting-failure-cases)
6. [Replacing Helm lint](#6-replacing-helm-lint)
7. [Replacing YAML lint](#7-replacing-yaml-lint)
8. [Adding schema validation with kubeconform](#8-adding-schema-validation-with-kubeconform)
9. [Gap analysis](#10-gap-analysis)

---

## 1. Install the tools

### helm-unittest

As a Helm plugin (recommended for local development):

```bash
helm plugin install https://github.com/helm-unittest/helm-unittest
```

Or as a standalone binary:

```bash
# macOS
brew install helm-unittest

# Go install
go install github.com/helm-unittest/helm-unittest/cmd/helm_unittest@latest
```

### chart-testing

```bash
# macOS
brew install chart-testing

# Docker (no install needed in CI)
docker run --rm -v $(pwd):/charts quay.io/helmpack/chart-testing:latest ct lint
```

### kubeconform

```bash
# macOS
brew install kubeconform

# Go install
go install github.com/yannh/kubeconform/cmd/kubeconform@latest
```

---

## 2. Concept mapping

| Hull concept | Equivalent |
|---|---|
| `test.Suite` | A directory of `*_test.yaml` files under `tests/` in your chart |
| `test.Case` (one per values config) | One `suite:` block per `*_test.yaml` file, with `set:` overrides |
| `test.NamedCheck` | An `it:` block inside a `suite:` — reuse by factoring into shared fixture files |
| `test.FailureCase` | `test/unit` suite with `failedTemplate: true` |
| `chart.NewTemplateOptions(...).SetValue(...)` | `set:` key under a test suite definition |
| `checker.PerResource(...)` assertions | `asserts:` entries in an `it:` block |
| `HelmLint` automatic check | `ct lint` or `helm lint --strict` |
| `YAMLLint` check | `ct lint` (has built-in yamllint support) or run yamllint directly |
| Manifest schema validation (implicit via Helm) | `kubeconform` |
| `coverage.Tracker` | No equivalent — manual tracking only |

---

## 3. Converting test cases

Hull test cases each use a `TemplateOptions` with `.SetValue()` calls. In helm-unittest, each test suite file represents one scenario, and values are overridden inline with `set:`.

### Hull (Go)

```go
Cases: []test.Case{
    {
        Name: "Using Defaults",
        TemplateOptions: chart.NewTemplateOptions("my-chart", "default"),
    },
    {
        Name: "Custom image tag",
        TemplateOptions: chart.NewTemplateOptions("my-chart", "default").
            SetValue("image.tag", "v1.2.3"),
    },
    {
        Name: "Custom data map",
        TemplateOptions: chart.NewTemplateOptions("my-chart", "default").
            Set("data", map[string]string{"hello": "cattle"}),
    },
},
```

### helm-unittest equivalent

Create one file per meaningful scenario under `tests/` in your chart directory:

**`tests/defaults_test.yaml`**

```yaml
suite: Using Defaults
templates:
  - templates/configmap.yaml
  - templates/deployment.yaml
tests:
  - it: renders with default values
    asserts:
      - isKind:
          of: ConfigMap
```

**`tests/custom_image_test.yaml`**

```yaml
suite: Custom image tag
templates:
  - templates/deployment.yaml
tests:
  - it: uses the overridden image tag
    set:
      image.tag: v1.2.3
    asserts:
      - equal:
          path: spec.template.spec.containers[0].image
          value: myrepo/myimage:v1.2.3
```

**`tests/custom_data_test.yaml`**

```yaml
suite: Custom data map
templates:
  - templates/configmap.yaml
tests:
  - it: includes overridden data
    set:
      data:
        hello: cattle
    asserts:
      - equal:
          path: data.hello
          value: cattle
```

Run all tests:

```bash
helm unittest ./my-chart
```

---

## 4. Converting named checks

Hull's `NamedCheck` runs the same assertion logic across every test case. In helm-unittest the most practical approach is to repeat assertions per file, or extract shared assertions using `templates:` scoping.

### Hull (Go)

```go
NamedChecks: []test.NamedCheck{
    {
        Name:   "ConfigMaps have expected data key",
        Covers: []string{".Values.data"},
        Checks: test.Checks{
            checker.PerResource(func(tc *checker.TestContext, cm *corev1.ConfigMap) {
                assert.Contains(tc.T, cm.Data, "config")
            }),
        },
    },
    {
        Name: "All workloads have a service account",
        Checks: test.Checks{
            checker.PerWorkload(func(tc *checker.TestContext, obj metav1.Object, pts corev1.PodTemplateSpec) {
                assert.NotEmpty(tc.T, pts.Spec.ServiceAccountName)
            }),
        },
    },
},
```

### helm-unittest equivalent

There is no single "run this check across all cases" mechanism. The two practical approaches are:

**Option A — inline assertions in every suite file (most explicit)**

Add the repeated assertion to each `tests/` file:

```yaml
# tests/defaults_test.yaml
suite: Using Defaults
templates:
  - templates/configmap.yaml
tests:
  - it: configmap has config key
    asserts:
      - isNotNull:
          path: data.config
```

```yaml
# tests/custom_data_test.yaml
suite: Custom data map
templates:
  - templates/configmap.yaml
tests:
  - it: configmap has config key
    set:
      data:
        hello: cattle
    asserts:
      - isNotNull:
          path: data.config
```

**Option B — one dedicated "invariants" test file per template**

Create a test file that verifies rules that must hold regardless of values, using `set:` to cover the important value permutations:

```yaml
# tests/invariants_configmap_test.yaml
suite: ConfigMap invariants
templates:
  - templates/configmap.yaml
tests:
  - it: always has config key with defaults
    asserts:
      - isNotNull:
          path: data.config

  - it: always has config key with custom data
    set:
      data:
        hello: cattle
    asserts:
      - isNotNull:
          path: data.config
```

```yaml
# tests/invariants_workloads_test.yaml
suite: Workload invariants
templates:
  - templates/deployment.yaml
tests:
  - it: deployment has a service account
    asserts:
      - isNotEmpty:
          path: spec.template.spec.serviceAccountName

  - it: deployment has a service account with custom values
    set:
      serviceAccount.name: custom-sa
    asserts:
      - equal:
          path: spec.template.spec.serviceAccountName
          value: custom-sa
```

### Common assertion types in helm-unittest

| Hull assertion | helm-unittest `asserts:` entry |
|---|---|
| `assert.Equal(tc.T, expected, actual)` | `equal: { path: ..., value: ... }` |
| `assert.Contains(tc.T, map, key)` | `isNotNull: { path: ... }` |
| `assert.NotEmpty(tc.T, val)` | `isNotEmpty: { path: ... }` |
| `assert.Nil(tc.T, val)` | `isNull: { path: ... }` |
| resource is of expected kind | `isKind: { of: Deployment }` |
| resource count | `hasDocuments: { count: 2 }` |
| label present | `matchRegex: { path: metadata.labels.app, pattern: .+ }` |
| labels equal | `equal: { path: metadata.labels, value: { app: myapp } }` |
| annotation value | `equal: { path: metadata.annotations["my/annotation"], value: "true" }` |

Full assertion reference: https://helm-unittest.github.io/helm-unittest/docs/assertion-types

---

## 5. Converting failure cases

Hull's `FailureCase` tests scenarios where chart rendering is expected to fail and asserts on the error message.

### Hull (Go)

```go
FailureCases: []test.FailureCase{
    {
        Name: "Set .Values.shouldFail",
        TemplateOptions: chart.NewTemplateOptions("my-chart", "default").
            SetValue("shouldFail", "true"),
        Covers:         []string{".Values.shouldFail"},
        FailureMessage: ".Values.shouldFail is set to true",
    },
    {
        Name: "Missing required value",
        TemplateOptions: chart.NewTemplateOptions("my-chart", "default").
            SetValue("image.repository", ""),
        FailureMessage: "image.repository is required",
    },
},
```

### helm-unittest equivalent

Use `failedTemplate: true` with an optional `errorMessage:` regex to match:

```yaml
# tests/failure_cases_test.yaml
suite: Failure cases
templates:
  - templates/configmap.yaml
tests:
  - it: fails when shouldFail is true
    set:
      shouldFail: "true"
    asserts:
      - failedTemplate:
          errorMessage: .Values.shouldFail is set to true

  - it: fails when image.repository is empty
    set:
      image.repository: ""
    asserts:
      - failedTemplate:
          errorMessage: image.repository is required
```

`errorMessage` is matched as a substring against the rendered error, so you do not need to match the full message.

---

## 6. Replacing Helm lint

Hull runs `helm lint --strict` automatically for every test case. Replace this with `ct lint` (which wraps `helm lint`) or run `helm lint` directly.

### `ct lint` (recommended)

chart-testing lint validates the chart against configurable rules and calls `helm lint --strict` under the hood. It also checks `Chart.yaml` fields, owner annotations, and version bumps.

```bash
# Lint a single chart
ct lint --charts ./my-chart

# Lint all charts that changed relative to main
ct lint --target-branch main
```

Add a `ct.yaml` config at the repo root to match any Rancher annotation requirements Hull was enforcing:

```yaml
# ct.yaml
chart-dirs:
  - charts
validate-maintainers: false
helm-extra-args: "--timeout 600s"
```

### Standalone `helm lint`

```bash
helm lint ./my-chart --strict
helm lint ./my-chart --strict --set image.tag=v1.2.3
```

### Replacing Rancher-specific annotation linting

Hull validates several `catalog.cattle.io/` annotations automatically. If you need to keep this, add a simple shell script or a `ct` custom lint rule. The annotations Hull checks are:

| Annotation | Rule |
|---|---|
| `catalog.cattle.io/display-name` | Must be present |
| `catalog.cattle.io/namespace` | Must be present |
| `catalog.cattle.io/release-name` | Must be present |
| `catalog.cattle.io/kube-version` | Must be a valid semver constraint |
| `catalog.cattle.io/rancher-version` | Must be a valid semver constraint |
| `catalog.cattle.io/permits-os` | Must be one of `linux`, `windows`, `linux,windows` |

A minimal check script:

```bash
#!/usr/bin/env bash
# scripts/lint-rancher-annotations.sh
set -euo pipefail

CHART_YAML="${1}/Chart.yaml"

check_annotation() {
  local key="$1"
  if ! grep -q "^  ${key}:" "$CHART_YAML"; then
    echo "ERROR: missing annotation: ${key}" >&2
    exit 1
  fi
}

check_annotation "catalog.cattle.io/display-name"
check_annotation "catalog.cattle.io/namespace"
check_annotation "catalog.cattle.io/release-name"
check_annotation "catalog.cattle.io/kube-version"
check_annotation "catalog.cattle.io/rancher-version"
check_annotation "catalog.cattle.io/permits-os"

permits_os=$(grep "catalog.cattle.io/permits-os" "$CHART_YAML" | awk -F': ' '{print $2}' | tr -d '"')
case "$permits_os" in
  linux|windows|"linux,windows"|"windows,linux") ;;
  *) echo "ERROR: catalog.cattle.io/permits-os has invalid value: ${permits_os}" >&2; exit 1 ;;
esac

echo "Rancher annotations OK"
```

---

## 7. Replacing YAML lint

Hull calls `yamllint` on rendered templates when `opts.YAMLLint.Enabled = true`. chart-testing includes yamllint integration out of the box.

### via `ct lint`

chart-testing calls yamllint automatically if it is installed. Configure its rules in a `.yamllint.yaml` at the repo root, matching Hull's default config:

```yaml
# .yamllint.yaml
rules:
  braces:
    min-spaces-inside: 0
    max-spaces-inside: 0
  colons:
    max-spaces-before: 0
  indentation:
    spaces: consistent
    indent-sequences: whatever
  key-duplicates: enable
  line-length: disable
  trailing-spaces: disable
  truthy:
    level: warning
```

### Standalone yamllint on rendered templates

```bash
helm template my-release ./my-chart | yamllint -c .yamllint.yaml -
```

---

## 8. Adding schema validation with kubeconform

Hull renders manifests and Helm validates them loosely. kubeconform validates rendered manifests against the official Kubernetes JSON schemas, catching invalid field names, wrong types, and missing required fields that Helm will not catch.

### Basic usage

```bash
# Render and pipe directly to kubeconform
helm template my-release ./my-chart | kubeconform -strict -summary

# Specify a Kubernetes version to validate against
helm template my-release ./my-chart | kubeconform -strict -summary \
  -kubernetes-version 1.28.0

# Include CRD schemas from a registry
helm template my-release ./my-chart | kubeconform -strict -summary \
  -schema-location default \
  -schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'
```

### Integrate into a Makefile target

```makefile
CHART_PATH ?= ./charts/my-chart
RELEASE_NAME ?= my-release
NAMESPACE ?= default
K8S_VERSION ?= 1.28.0

.PHONY: validate
validate:
	helm template $(RELEASE_NAME) $(CHART_PATH) \
	  --namespace $(NAMESPACE) \
	| kubeconform \
	    -strict \
	    -summary \
	    -kubernetes-version $(K8S_VERSION)
```

### Validate multiple value configurations

```bash
for values_file in tests/values/*.yaml; do
  echo "--- validating with $values_file ---"
  helm template my-release ./my-chart -f "$values_file" \
    | kubeconform -strict -summary -kubernetes-version 1.28.0
done
```

---

## 9. Gap analysis

Features Hull provides that have no direct replacement in this stack:

| Hull feature | Status | Notes |
|---|---|---|
| Test case matrix (multiple value configs) | Covered | helm-unittest `set:` per test |
| Per-resource assertions | Covered | helm-unittest `asserts:` |
| Failure case testing | Covered | helm-unittest `failedTemplate:` |
| Helm lint | Covered | `ct lint` / `helm lint --strict` |
| YAML lint | Covered | `ct lint` + `.yamllint.yaml` |
| Schema validation | Covered (improved) | kubeconform with versioned schemas |
| Rancher annotation linting | Partial | Needs a small shell script |
| `checker.PerWorkload` multi-resource logic | Partial | helm-unittest does not support cross-resource assertions in one test; split into separate tests or use a shell script that validates rendered YAML |
| `checker.Store` / cross-check state | Not covered | helm-unittest tests are stateless per assertion |
| Coverage tracking | Not covered | No equivalent; consider maintaining a manual checklist in your test directory |
| Go-native assertions (`testify`) | Not covered | helm-unittest assertions are YAML-based; complex business logic requires scripted validation outside helm-unittest |

For charts with complex cross-resource validation (e.g., "every workload must reference a ServiceAccount that exists in this chart"), a lightweight script that validates the rendered manifest bundle is the pragmatic replacement:

```bash
#!/usr/bin/env bash
# scripts/validate-service-accounts.sh
# Checks that all workload serviceAccountName references exist as ServiceAccount resources.

set -euo pipefail

MANIFESTS=$(helm template test-release "$1")

# Extract all ServiceAccount names
sa_names=$(echo "$MANIFESTS" | yq e 'select(.kind == "ServiceAccount") | .metadata.name' -)

# Extract all serviceAccountName references from workloads
refs=$(echo "$MANIFESTS" \
  | yq e 'select(.kind == "Deployment" or .kind == "DaemonSet" or .kind == "StatefulSet") | .spec.template.spec.serviceAccountName' - \
  | grep -v "^null$")

while IFS= read -r ref; do
  if ! echo "$sa_names" | grep -qx "$ref"; then
    echo "ERROR: workload references ServiceAccount '$ref' which does not exist in the chart" >&2
    exit 1
  fi
done <<< "$refs"

echo "Service account references OK"
```
