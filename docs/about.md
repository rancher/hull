# About Hull

## What is Hull?

Hull is a **Go testing framework** for writing comprehensive tests on [Helm](https://github.com/helm/helm) charts.

## How Hull Works

Hull, as a testing framework, could be simply described as something analagous to "API Testing" framework on the `values.yaml` that is used to configure a Helm chart.

### What Does Hull Test?

When a Helm chart owner defines a Helm chart, they usually define a base set of manifests that need to be deployed under the `templates/` directory.

Then they convert those manifests into Go templates that are rendered based on Helm's [Built-In `RenderValues` struct](https://helm.sh/docs/chart_template_guide/builtin_objects/).

So when you render a Helm chart (for a `helm template`, `helm install`, `helm upgrade`, etc.), what Helm is really doing under the hood is taking your command line arguments to create a `RenderValues` struct and **supplying that struct to generate a Kubernetes manifest**.

Therefore, from a testing perspective, what we would ideally like to do to set up comprehensive unit testing is **test your Helm chart against all possible values of the `RenderValues` struct** (or at least as many as is reasonable to encode in Go tests).

### A Simple Introduction To Hull

Let's take a look at the anatomy of a simple Hull testing file:

```go
var suite = test.Suite{
	ChartPath:    filepath.Join("..", "testdata", "charts", "no-schema"),

	Cases: []test.Case{
		{
			Name: "Setting .Values.Data",

			TemplateOptions: chart.NewTemplateOptions("no-schema", "default").
				SetValue("data.hello", "world"),

			Checks: []checker.Check{
				{
					Name: "sets .data.config on ConfigMaps",
					Func: workloads.EnsureConfigMapsHaveData(
						map[string]string{"config": "hello: world"},
					),
				},
			},
		},
	},
}

func TestChart(t *testing.T) {
	suite.Run(t, nil)
}
```

Here, we're defining a `test.Suite` that applies to the chart located at [`../testdata/charts/no-schema`](../testdata/charts/no-schema).

Within the `test.Suite`, we define a list of `test.Case` objects that will be applied to the chart; each case takes in a set of `TemplateOptions`, which effectively allows you to configure the underlying `RenderValues` struct that would be passed into Helm when it tries to render your template.

For example, the template that Hull would render on parsing that `test.Case` would be equivalent to the one produced by running `helm template no-schema -n default --set data.hello=world ../testdata/charts/no-schema` based on the `TemplateOptions` provided.

> **Note**: The utility function `SetValue` is intentionally written to mimic the way that Helm would receive these arguments itself (for the convenience of the tester)! However, you could also pass in a custom `TemplateOptions` struct for more complex cases (such as if you want to modify the `Capabilities` or `Release` options provided to the template call).
>
> See [`pkg/chart/template_options.go`](../pkg/chart/template_options.go) to view the `TemplateOptions` struct and see what options it takes!

#### Running a `test.Case` in Hull

On executing **each** `test.Case`, Hull automatically takes the following actions:
1. A `helm lint` will be run on the rendered template **(see note below)**
2. All the `suite.DefaultChecks` will be run on the template (a `[]checker.Checks` that are expected to be run on all templates)
3. All the `checker.Checks` in `suite.Cases[i].Checks` will be run on the generated template

> **Note**: Each `test.Case` is run as a separate [Go subtest](https://go.dev/blog/subtests), so it's possible to run individual cases for charts on a `go test`!

> **Note**: Hull adds additional linting for Rancher charts, such as validating the existence of certain annotations in the correct format. This can be enabled by supplying additional options in the second argument of the `suite.Run` call, but is disabled by default.
>
> To encode additional linting, Hull uses the same underlying mechanism as Helm does, as seen in [`pkg/chart/template_lint.go`](../pkg/chart/template_lint.go).
>
> Feature requests to add additional custom linters (or the ability to supply custom linters that organizations can "plug-in" to the `helm lint` action) are welcome!

> **Note**: If you see a lint failure and want to debug where it is coming from, Hull natively supports the advanced capability to **output a Markdown file** to a location identified by the environment variable `TEST_OUTPUT_DIR`.
>
> When this environment variable is set, Hull will create a file at `${TEST_OUTPUT_DIR}/test-${UNIX_TIMESTAMP}.md` **on every failed test execution** that formats all the tests errors in a human-readable way.

#### What Is A `checker.Check`?

A `checker.Check` is a definition of a single check you want to run on a generated template.

Examples of what you may want to encode in a `checker.Check` include:
- Resources follow best practices that allow them to be deployed onto specialized environments (i.e. Deployments have nodeSelectors and tolerations for Windows)
- Resources meet other business-specific requirements (i.e. all images start with `rancher/mirrored-` unless they belong on a special allow-list of Rancher original images)

This is the core of what a chart owner will want to be able to use Hull for; therefore, by design, Hull is **extremely permissive** with respect to how a chart developer can choose to write their checks!

To elaborate, you may notice that the `checker.Check` contains two fields: a name for the check, used to identify which check may have failed on a test failure, and a `CheckFunc`. However, if you look at the type definition of a `CheckFunc`, [you may notice that it is an empty `interface{}`](../pkg/checker/types.go)!

This is because, under the hood, Hull performs the validation of the `CheckFunc`'s function signature **at runtime** based on logic encoded in [`pkg/checker/internal/do.go`](../pkg/checker/internal/do.go).

> **Note**: The logic used for the checker is heavily tested in [`pkg/checker/internal/do.go`](../pkg/checker/internal/do_test.go) with multiple extremely complex examples. However, contributions for any testing gaps are welcome!

In essence, what Hull looks for in a `CheckFunc` is any object matches a signature of `func(*testing.T, someRuntimeObjectStruct)` where `someRuntimeObjectStruct` is defined as "any struct that contains only slices of objects that implement [`k8s.io/apimachinery/pkg/runtime.Object`](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#Object) **somewhere in the struct**".

For example, a function with a signature like `func(t *testing.T, cms struct{ []*corev1.ConfigMap })` is perfectly acceptable! Another perfectly acceptable function would be:

```go
type myStruct {
  CronJobs []*batchv1.CronJob
  DaemonSets []*appsv1.DaemonSet
  Deployments []*appsv1.Deployment
  Jobs []*batchv1.Job
  StatefulSets []*appsv1.StatefulSet
}

func MyCheck(t *testing.T, workloads myStruct) {

}
```

On introspecting on the function's signature, **Hull will automatically manage placing the applicable rendered template objects into your provided struct to execute each test!**

In this way, Hull encourages chart developers to think of changes applied to `values.yaml` as changes to **sub-manifests**: specific sets of resource types encoded in `someRuntimeObjectStruct` that you, the chart tester, cares about when writing the test.

For example, in the above `MyCheck` function that takes in all the workloads types, you may want to encode a check that determines whether all those workloads have resource requests and/or limits set. This would be fairly easy to do in Hull; just loop through each of the objects in your desired struct and execute the check, emitting a `t.Error` if it fails the check.

You may want to do something even more complex, such as encoding whether all the images in these workloads exist on DockerHub. While this would be difficult to do in purely YAML-based linting solutions, this is something that can be encoded in a `checker.Check` by importing the Golang Docker client and leveraging it in a custom `checker.CheckFunc` to make the calls.

#### Using A Built-In `checker.Check`

Ideally, the intention of Hull is to provide a full set of utility functions in [`pkg/checker/resource`](../pkg/checker/resource) that emit `checker.CheckFunc` functions that cover common types of checks that users will want to execute.

For example, in the [`examples/example_test.go`](../examples/example_test.go) the following utility function from [`pkg/checker/resource/workloads`](../pkg/checker/resource/workloads/) emits a `checker.CheckFunc` that ensure that the number of `ConfigMap` objects emitted is equivalent to the number provided:

```go
workloads.EnsureNumConfigMaps(2)
```

While the current set of utility functions is not comprehensive yet, contributions are welcome!

#### Writing A Custom `checker.Check`

If you plan to write your own `checker.CheckFunc`, the major caveat is that you need to ensure that any Kubernetes resource Go types are added to the `pkg/checker.Scheme` defined [here](../pkg/checker/checker.go).

For users who have developed Kubernetes controllers before, it may be familiar to mention that the usual way to do this is by running the `AddToScheme` function usually generated by most Go-based Kubernetes controller frameworks.

For example, here is how you would add the `corev1` resources to ensure the Hull will be able to correctly identify that a YAML object identified a particular `apiVersion` and `kind` should be marshalled into the `*corev1.ConfigMap` struct:

```go
import (
	"github.com/aiyengar2/hull/pkg/checker"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	if err := corev1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
}
```

If you also needed your `CheckFunc` to have workloads, you may also need to import `appsv1` and `batchv1`, so you can modify this accordingly:

```go
import (
	"github.com/aiyengar2/hull/pkg/checker"
	corev1 "k8s.io/api/core/v1"
  appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

func init() {
	if err := appsv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := batchv1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
}
```

And that's all you need to do!

Alternatively, it is highly recommended to define structs that simply embed the structs defined in [`pkg/checker/resource`](../pkg/checker/resource), which **already take care of these imports on `init()` for you.**

> **Note**: Why is this necessary?
>
> Without adding a type to the underlying `checker.Scheme`, Hull will not understand how to convert a YAML object it sees with a given `apiVersion` and `kind` to a corresponding Go struct that is within your function signature.
>
> If a type cannot be found in the `checker.Scheme`, it will **always** be assumed that your object is of type [`*unstructured.Unstructured`](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured#Unstructured).

### Defining Test "Coverage" In Hull

While we've talked about how to define tests in Hull so far, Hull can also determine the degress of **coverage** that your `test.Suite` is currently testing on your chart.

This generally involves three steps:
1. Defining a Go struct type representing your `values.yaml` (`ExampleChart`)
2. Adding an empty struct of that type to your `suite.ValuesStruct`
3. Calling `TestCoverage` in the way defined below

For example:

```go
type ExampleChart struct {
  // filled out by Chart Owner
}

var suite = test.Suite{
	...
	ValuesStruct: &ExampleChart{},
  ...
}

func TestCoverage(t *testing.T) {
	templateOptions := []*chart.TemplateOptions{}
	for _, c := range suite.Cases {
		templateOptions = append(templateOptions, c.TemplateOptions)
	}
	coverage, report := test.Coverage(t, ExampleChart{}, templateOptions...)
	if t.Failed() {
		return
	}
	assert.Equal(t, 1.00, coverage, report)
}
```

#### How Is Coverage Defined?

As described in [a previous section](#what-does-hull-test), Hull seeks to **test your Helm chart against all possible values of the `RenderValues` struct** (or at least as many as is reasonable to encode in Go tests).

However, when it comes to defining "coverage" today, Hull defines it more simply: **whether the set of `TemplateOptions` defined by each `test.Case` covers configuring every available field within the `values.yaml` at least once.**.

> **Note**: An open issue exists in https://github.com/aiyengar2/hull/issues/5 around how to improve these coverage calculations to get closer to the expected goal.

#### Calculating Current Coverage

Based on this simpler definition of Hull coverage, the current coverage is easier to calculate. Hull takes the following process:
- Convert each `TemplateOption` into a `map[string]interface{}` (i.e. JSON blob) representation and merge them all together into one blob (doesn't matter if there are overwrites)
- Keep track of every nested key that is set in the combined `map[string]interface{}`

> **Note**: While this may end up keeping track of **more** keys than necessary (i.e. you may store `deployments.labels.hello` as a key when only `deployments.labels` needs to be stored, since `hello` is just a random label selected for a test), it will never collect less. This will be important when calculating coverage.

#### Calculating Total Coverage

While the simpler definition of coverage seems more manageable to calculate total coverage, there's one fundamental problem: **Helm does not require `values.yaml` files to have a strict schema by default.**

While upstream Helm has introduced the **capability** for Helm chart owners to specify a [`values.schema.json`](https://helm.sh/docs/topics/charts/#schema-files) (a [JSON Schema](https://json-schema.org/) validated by the Helm chart on template render), it's not actively maintained by several charts.

However, since Hull requires this schema to understand how to calculate coverage and JSON schemas are hard to hand-maintain, the approach Hull has taken is to allow you to **describe your `values.yaml` representation in a Go struct type**.

As a result, Hull is able to leverage type introspection to identify all possible fields that need to be set on the type by using the following logic:
- If it is a slice: ensure the slice is non-nil / empty at least once
- If it is a map: ensure the map is non-nil / empty at least once
- If it is any other type: ensure the value is set at least once

In return for defining this struct, Hull will leverage [`invopop/jsonschema`](https://github.com/invopop/jsonschema) to automatically translate your Go struct (while respecting struct tags identified by `jsonschema:`!) into a JSON schema, which will automatically add or replace the existing `values.schema.json` file of the chart.

> **Note**: Automatic management of your `values.schema.json` can be disabled by setting `DoNotModifyChartSchemaInPlace` in the options provided to the `test.Suite` on `suite.Run`.

#### Calculating Coverage (%)

Putting it all together, Hull outputs the total coverage on calling `test.Coverage(t, ExampleChart{}, templateOptions...)` by counting the number of fields identified by **both** your current coverage and total coverage and dividing it by the number of fields identified by your total coverage.

On an error, Hull will also print out which fields are lacking tests. For example:

```log
--- FAIL: TestCoverage (0.00s)
    example_test.go:85:
                Error Trace:    example_test.go:85
                Error:          Not equal:
                                expected: 1
                                actual  : 0
                Test:           TestCoverage
                Messages:       The following keys are not set: [.data]
                                Only the following keys are covered: []
```

## Should I Use Hull?

### Is this the best Helm testing framework for me?

Depends on your use case!

Hull is great for organization that:
- Needs a fairly robust Helm template testing tool with easy extensibility
- Works primarily in Golang (such as those that develop Golang Kubernetes operators and ship them in Helm charts)

For example, Hull is targeted for use in [`rancher/charts`](https://github.com/rancher/charts/tree/main/charts) today, a repository that maintains a large number of highly complex charts.

### Alternatives

Depending on your organizational requirements, you may benefit from another Helm testing framework which might be simpler to use than Hull.

Here are some popular options that were considered on developing Hull:

| Project                                                               | Pros                                                                                                                                                                                                                                                                                             | Cons                                                                                                                                                                                                                                                                                     | Recommended If...                                                                                                                                                                                                                                                                                                                                                             |
|-----------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [`stackrox/kube-linter`](https://github.com/stackrox/kube-linter)     | Lots of great default checks out of the box!   <br /><br />  Easy to add some forms of custom checks in YAML                                                                                                                                                                                     | Adding a new custom check requires forking kube-linter and building it yourself                                                                                                                                                                                                          | You would like a simpler out-of-the-box solution around testing your Helm manifests to see if they follow best practices and don't anticipate adding many custom checks.  <br /> <br />  Might be great to leverage this along with [`yannh/kubeconform`](https://github.com/yannh/kubeconform) for resource schema validation.                                               |
| [`gruntwork-io/terratest`](https://github.com/gruntwork-io/terratest) | Go-based framework is easy to extend                                                                                                                                                                                                                                                             | Not a "Helm-first" framework; great from an integration testing perspective for performing helm operations on a live cluster, but does not have much around marshalling and unmarshalling resources in rendered template manifests (unless you have one resource per Helm template file) | You would like more of an integration testing solution, no need for **robust** template-based validation like what Hull offers.  <br /> <br />  Might be a great idea to consider using this, or a simpler solution like [`helm/chart-testing`](https://github.com/helm/chart-testing) if all you want is to install / upgrade and run `helm test` on the basic configuration |
| [`conftest`](https://www.conftest.dev/)                               | Ability to define tests in the same language as OPA, which is great for asserting policies on manifests generated from Helm charts  <br /><br />  Possibility to utilize currently available policy libraries to maintain a common set of policies for chart best practices and policy execution | Rego is hard to learn, use, and debug for developers who don't directly work with it, as compared to using more common languages like Python or Go                                                                                                                                       | Your organization has familiarity with using Rego as a policy language.                                                                                                                                                                                                                                                                                                       |
| [`quintush/helm-unittest`](https://github.com/quintush/helm-unittest) | Ability to define tests in pure YAML, which is simpler to encode   <br /><br />  Framework is designed to be BDD-style                                                                                                                                                                           | YAML is hard to extend, especially for complicated sets of tests                                                                                                                                                                                                                         | Your organization would like a simpler out-of-the-box solution that can be maintained without needing to understand and maintain code in another programming language                                                                                                                                                                                                         |

> **Note**: The above pros and cons may not match the current state of these projects. We welcome contributions to add additional pros and cons if any of these projects may have been misrepresented!
