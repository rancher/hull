# About Hull

## What is Hull?

Hull is a **Go testing framework** for writing comprehensive tests on [Helm](https://github.com/helm/helm) charts.

### How Do Helm Charts Work?

At its core, all Helm charts are just sets of [Go templates](https://pkg.go.dev/text/template) listed under a `templates/` directory that are rendered based on Helm's [Built-In `RenderValues` struct](https://helm.sh/docs/chart_template_guide/builtin_objects/), which takes input from the `values.yaml` file declared with the chart.

> **Note**: Helm places the chart's metadata (i.e. the `Chart.yaml`) under `.Chart` and the `values.yaml` under `.Values`. For subcharts, it places all overrides provided by the main chart under `.Values.<subchart>`.

Upon rendering, it's expected that each file produces a **Kubernetes manifest**, or a list of Kubernetes resources (where the type of resource for each [YAML document](https://www.yaml.info/learn/document.html) produced is identified by the `apiVersion` and `kind` fields).

> **Note**: The combined list of all resources in the Kubernetes manifests produced by all template files in a chart constitute **a single Helm release**.

For example, take a simple Helm chart that has this single file in its `templates/` directory:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: {{ .Values.image }}
    ports:
    - containerPort: 80
```

By providing a `values.yaml` that contains `image: nginx:latest` to this Helm chart or running the `helm install` or `helm upgrade` command with `--set image=nginx:latest`, Helm will produce the following manifest:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:latest
    ports:
    - containerPort: 80
```

If you were to provide `image: nginx:1.2.3` instead, you would get a different template:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.2.3
    ports:
    - containerPort: 80
```

If you are specifically running a `helm install` (non-dry-run) or `helm upgrade`, Helm will then take this produced manifest and do the equivalent of a `kubectl apply` onto the cluster.

> **Note**: It's not exactly correct to say that Helm does a `kubectl apply` since Helm supports [Chart Hooks](https://helm.sh/docs/topics/charts_hooks/), each of which applies resources in order based on the weight of the hook annotations attached to it. A normal `kubectl apply` would apply all resources in a given manifest in one shot.

### How Does Hull Support Testing Helm Charts?

Hull seeks to enforce comprehensive unit testing by helping you:
1. **Encode `test.Case`s that cover all possible valid configurations of the chart**: each `test.Case` corresponds to a single Helm template / release. This means each `test.Case` is tied to a single generated Kubernetes manifest rendered based on a given `values.yaml` (provided via `chart.TemplateOptions`) and the templates contained within the chart
2. **Run "checks" on all generated Helm releases**: these should be generic tests (i.e. `test.NamedCheck`s) that run on all templates that are generated from the `test.Case`s. Thes tests can be parametrized by the `values.yaml` that was used to render a given `chart.Template` via using the `checker.RenderValue / checker.MustRenderValue` helper functions

### Testing With and Without Hull

Let's say you wanted to manually test the `.Values.image` field above without Hull.

The intended behavior of the field is to override the contents of `.spec.containers[0].image` to whatever value is provided.

Therefore, there are probably two checks you would write up here:
1. `helm template <default-chart-release-name> -n <default-chart-namespace> | yq e 'select(.kind == "Pod") | .spec.containers[0].image'`: ensure that the output here is the default value of `nginx:latest`
2. `helm template <default-chart-release-name> -n <default-chart-namespace> --set image=nginx:1.2.3 | yq e 'select(.kind == "Pod") | .spec.containers[0].image'`: ensure that the output here is the provided value `nginx:1.2.3`

For a simple field like this, these two checks would be sufficient to fully test this field.

On the other hand, in Hull you would encode this as two `test.Case`s with two distinct `test.NamedChecks` that run on produced Helm templates from those options:

```go
var ChartPath = utils.MustGetPathFromModuleRoot("path", "to", "chart")

var (
	DefaultReleaseName = "<default-chart-release-name>"
	DefaultNamespace   = "<default-chart-namespace>"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name: "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Setting nginx:1.2.3 image",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("nginx.image", "1.2.3"),
		},
	},

	NamedChecks: []test.NamedChecks{
		{
			Name: "Has correct value for Pod's image"
			Covers: []string{".Values.nginx.image"}
			Checks: test.Checks{
				// TODO: some check(s) that ensure .spec.containers[0].image for the nginx workload == .Values.nginx.image
			}
		}
	},
}


func TestChart(t *testing.T) {
	opts := test.GetRancherOptions()
	suite.Run(t, opts)
}
```

While Hull may seem more verbose, there are three distinct advantages of encoding those same manual checks into Hull:
1. It runs along with your normal CI; by just enforcing that a `go test -count=1 <testing-module>` passes in your repository, you know that all written checks would pass on all recorded `test.Case`s that are used with your chart, so a change that has been introduced to the chart is not breaking an existing use case
2. You can encode checks that span across all of the templates by only declaring a single `test.NamedCheck`; for example, a check that sees that all Pods generated have `nodeSelectors` and `tolerations` to support being deployed in a cluster with Windows nodes would apply to all `values.yaml` configurations encoded in `test.Case`s
3. Hull's coverage check will ensure you have at least one test case to cover all used fields from the `values.yaml`, so you'll be able to use it to identify any gaps in CI

> **Note**: Why do we run the `go test` with `-count=1`?
>
> Go normally will cache the results of tests to avoid re-running them when the underlying Go code has not been modified.
>
> However, this would not invalidate the cache **when your Helm chart alone changes**; as a result, passing in `count=1` tells `go test` to always run every test at least once, ignoring the cached contents.

### A Simple Introduction To Hull

Let's take a look at the anatomy of a simple Hull testing file:

```go
var ChartPath = utils.MustGetPathFromModuleRoot("..", "testdata", "charts", "simple-chart")

var (
	DefaultReleaseName = "simple-chart"
	DefaultNamespace   = "default"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
	},
}

func TestChart(t *testing.T) {
	suite.Run(t, nil)
}
```

Here, we're defining a `test.Suite` that applies to the chart located at [`../testdata/charts/simple-chart`](../testdata/charts/simple-chart).

On running `go test -count=1 -v <path-to-module>`, we get the following output:

```log
=== RUN   TestChart
=== RUN   TestChart/Using_Defaults
=== RUN   TestChart/Using_Defaults/HelmLint
    template.go:171: [INFO] Chart.yaml: icon is recommended
=== RUN   TestChart/Coverage
    suite.go:208:
                Error Trace:    path/to/hull/examples/tests/workspace/suite.go:208
                Error:          Not equal:
                                expected: 1
                                actual  : 0
                Test:           TestChart/Coverage
                Messages:       The following field references are not tested:
                                - {{ .Values.data }} : templates/configmap.yaml
                                - {{ .Values.shouldFail }} : templates/configmap.yaml
                                - {{ .Values.shouldFailRequired }} : templates/configmap.yaml
--- FAIL: TestChart (0.01s)
    --- PASS: TestChart/Using_Defaults (0.01s)
        --- PASS: TestChart/Using_Defaults/HelmLint (0.01s)
    --- FAIL: TestChart/Coverage (0.00s)
FAIL
FAIL    path/to/hull/examples/tests/workspace      1.643s
FAIL
```

> **Note**: If you look at the output above, you'll notice that each `test.Case` is run as a separate [Go subtest](https://go.dev/blog/subtests) and tests within it are run as subtests of itself (i.e. `HelmLint`), so it's possible to run individual cases for charts on a `go test`!

The reason why you get this error is that, if you look at the contents of [`testdata/charts/simple-chart/templates/configmap.yaml`](../testdata/charts/simple-chart/templates/configmap.yaml), you will find that your Go template uses `{{ .Values.data }}` to populate the contents of the ConfigMaps deployed alongside the chart. It also uses `{{ .Values.shouldFail }}` and `{{ .Values.shouldFailRequired }}` for other custom logic.

These field usages are picked up by the logic in [`pkg/tpl`](../pkg/tpl/), which analyzes each file in the Helm chart provided to `suite.ChartPath` and automatically figures out the ground that needs to be covered to fully test the chart.

To resolve these issues, you will need to add an `test.Case` to this example suite that uses each of these fields and include a `test.NamedCheck`s for each of those fields to run the logical checks.

Alternatively, you can write a `test.FailureCase`, if the chart is expected not to render when some configuration is provided; this is used to test `.Vales.shouldFail` and `.Values.shouldFailRequired`.

### Adding a `test.Case`

To resolve the issue that came up as part of coverage, we first need to introduce a new test case that overrides `.Values.data`. 

To do this, we will first add a `test.Case`, which makes our `test.Suite` object look as follows:

```go
var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name:            "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("data.hello", "cattle"),
		},
	},
}
```

As you can see above, the `chart.TemplateOptions` provided differ between these two cases; the template that Hull would render on parsing the second `test.Case` would be equivalent to the one produced by running `helm template simple-chart -n default --set data.hello=cattle ../testdata/charts/simple-chart` based on the `TemplateOptions` provided.

It is also identical to provide the same configuration in the following way:

```go
var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{
					"hello": "cattle",
				}),
		},
	},
}
```

Or even like this:

```go
type MyStruct{
	Hello string
}

var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", MyStruct{
					Hello: "cattle",
				}),
		},
	},
}
```

Regardless of the object provided, the argument to `Set` will always marshall the given content into JSON and pass it to Helm as if it was received by the `--set-json` command line option.

Therefore, the second `test.Case` in both of the examples above would be equivalent to the one produced by running `helm template simple-chart -n default --set-json '{"data": {"hello": "cattle"}}' ../testdata/charts/simple-chart` based on the `TemplateOptions` provided.


> **Note**: `chart.TemplateOptions` allows you to configure the underlying `RenderValues` struct that would be passed into Helm when it tries to render your template; it corresponds directly to all flags providable in a `helm install`, `helm upgrade`, or `helm template` call that affects rendering.
>
> The utility functions `Set` and `SetValue` are intentionally written to mimic the way that Helm would receive these arguments itself (for the convenience of the tester)! However, you could also pass in a custom `TemplateOptions` struct for more complex cases (such as if you want to modify the `Capabilities` or `Release` options provided to the template call).
>
> See [`pkg/chart/template_options.go`](../pkg/chart/template_options.go) to view the `TemplateOptions` struct and see what options it takes!

#### Running a `test.Case` in Hull

On executing **each** `test.Case`, Hull automatically takes the following actions:
1. A `helm lint` will be run on the rendered Kubernetes manifest
2. Each `suite.NamedChecks` that does not exist in `case.OmitNamedChecks` will be run on the rendered Kubernetes manifests
3. Coverage will be checked; if the chart is not fully covered by the `test.Suite`, the test will fail

> **Note**: Since our current suite has no `test.NamedCheck`s, **no checks will be run on the template yet**.
>
> Therefore, you will only see 6 tests run: the root test for the overall chart, coverage for the overall chart, and 2 tests per case corresponding to the root test and just the output from running `helm lint`.
>
> Once we add checks, you will see 1 additional subtest per `test.Case` per `test.NamedCheck` be added to the total number of tests that are run.

We can see that tests for our new `test.Case`s have been added to our output on running `go test -count=1 -v <path-to-module>`

```log
=== RUN   TestChart
=== RUN   TestChart/Using_Defaults
=== RUN   TestChart/Using_Defaults/HelmLint
    template.go:171: [INFO] Chart.yaml: icon is recommended
=== RUN   TestChart/Override_.Values.data
=== RUN   TestChart/Override_.Values.data/HelmLint
    template.go:171: [INFO] Chart.yaml: icon is recommended
=== RUN   TestChart/Coverage
    suite.go:208:
                Error Trace:    path/to/hull/examples/tests/workspace/suite.go:208
                Error:          Not equal:
                                expected: 1
                                actual  : 0
                Test:           TestChart/Coverage
                Messages:       The following field references are not tested:
                                - {{ .Values.data }} : templates/configmap.yaml
                                - {{ .Values.shouldFail }} : templates/configmap.yaml
                                - {{ .Values.shouldFailRequired }} : templates/configmap.yaml
--- FAIL: TestChart (0.01s)
    --- PASS: TestChart/Using_Defaults (0.01s)
        --- PASS: TestChart/Using_Defaults/HelmLint (0.01s)
    --- PASS: TestChart/Override_.Values.data (0.00s)
        --- PASS: TestChart/Override_.Values.data/HelmLint (0.00s)
    --- FAIL: TestChart/Coverage (0.00s)
FAIL
FAIL   path/to/hull/examples/tests/workspace      0.845s
FAIL
```

> **Note**: Hull adds additional linting for Rancher charts, such as validating the existence of certain annotations in the correct format. This can be enabled by supplying additional options in the second argument of the `suite.Run` call, but is disabled by default.
>
> To encode additional linting, Hull uses the same underlying mechanism as Helm does, as seen in [`pkg/chart/template_lint.go`](../pkg/chart/template_lint.go).
>
> Feature requests to add additional custom linters (or the ability to supply custom linters that organizations can "plug-in" to the `helm lint` action) are welcome!

> **Note**: If you set `suiteOptions.YAMLLint.Enabled` to true (default is false), Hull will also run [`yamllint`](https://github.com/adrienverge/yamllint) on the YAML generated for each `test.Case` based on the configuration in [pkg/chart/configuration/yamllint.yaml](../pkg/chart/configuration/yamllint.yaml) (or whatever you provide to `suite.YAMLLint.Configuration`). 
>
> However, this does cause Hull to have an external dependency as you will need to have `yamllint` installed on your machine (or in the container you are using to run Hul) to run this check.

> **Note**: If you see a lint failure and want to debug where it is coming from, Hull natively supports the advanced capability to **output a Markdown file** to a location identified by the environment variable `TEST_OUTPUT_DIR`.
>
> When this environment variable is set, Hull will create a file at `${TEST_OUTPUT_DIR}/test-${UNIX_TIMESTAMP}.md` **on every failed test execution** that formats all the tests errors in a human-readable way.

Our next step is to add a check!

#### Adding a `test.NamedCheck`

In order to ensure that we can pass coverage for `.Values.data`, we will introduce a no-op `test.NamedCheck` to our test suite. 

On adding it, our test suite should look as follows:

```go
var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{
					"hello": "cattle",
				}),
		},
	},
	NamedChecks: []test.NamedCheck{
		{
			Name:   "Test",
			Covers: []string{".Values.data"},
			Checks: test.Checks{},
		},
	},
}
```

If you were to modify the suite to match what has been modified above, **you will find that .`Values.data` is passing coverage now!**

This is because there exists at least one `test.Case` whose TemplateOptions modify `.Values.data` in some way **and** at least one `test.NamedCheck` that runs against that `test.Case` that covers that particular value; as long as this requirement is satisfied, Hull is happy to let coverage pass for that field.

#### What are `test.Checks`?

While it's great that tests are passing in our example above, we're still passing tests as a false positive here; we need to actually execute a check on the manifest that is generated to truly have covered this field of the chart.

To do this, you can specify `test.Checks`, which is a slice of `checker.ChainedCheck`s; each `checker.ChainedCheck` produces a function that runs on the Kubernetes objects contained in a rendered Kubernetes manifest generated from a Helm chart.

Each `chart.Template` can run the `test.Checks` by running `template.Check(testingT, checker.NewCheckFunc(checks))`; this is precisely what the `test.Suite` does on a `Run`, with a couple of extra configuration options.

#### What should a `checker.ChainedCheck` do?

Examples of what you may want to encode in a `checker.ChainedCheck` function include:
- Resources follow best practices that allow them to be deployed onto specialized environments (i.e. Deployments have nodeSelectors and tolerations for Windows)
- Resources meet other business-specific requirements (i.e. all images start with `rancher/mirrored-` unless they belong on a special allow-list of Rancher original images)

This is the core of what a chart owner will want to be able to use Hull for; therefore, by design, Hull is **extremely permissive** with respect to how a chart developer can choose to write their checks!

#### Writing a `checker.ChainedCheck`

In order to create a `checker.ChainedCheck`, you can use any of the helper functions defined in [`pkg/checker/loop.go`](../pkg/checker/loop.go) (or any of the other files in the [`checker`](../pkg/checker/) module); these helper functions are intended to simplify the workflow of defining a `checker.ChainedCheck` by allowing a user to provide functions based on [Go Generics](https://go.dev/doc/tutorial/generics) to do common actions on resources.

All `checker.ChainedChecks` (including those helper functions) will always have a function that takes in the `checker.TestContext`, a construct that allows checks to perform contextual actions, such as:
- Running `checker.RenderValue[myType](tc, ".Values.something.i.want")` to extract a value from the rendered values of the `chart.Template`, such as `.Chart.Name` or `.Values.data`
- Running `checker.Store("Some Value I Store", "myValue")` / `checker.Get[string, string]("Some Value I Store")` to set arbitrary values for future checks in the chain to use
- Running `checker.MapSet` / `checker.MapGet` to set groupings that you can later iterate through using `checker.MapFor`; this is useful if you want to do something like find all workloads running Windows workloads (`.MapSet`) in one check and check if they have `nodeSelectors` set for Windows (`.MapFor`) in another
- Running `checker.HasLabels(obj, expectedLabels)` or `checker.HasAnnotations(obj, expectedLabels)` to simplify things that would need to be done for any arbitrary `metav1.Object`; **contributions are welcome for more such functions!**
- Running `checker.ToYAML` or other functions that handle performing basic transformations from objects to string representations for you

#### Writing a custom `checker.ChainedCheck`

You can even define your own `checker.ChainedCheck` in Hull, but to do this you will need to define a function that takes in a `checker.TestContext` and returns a `CheckFunc`. However, if you look at the type definition of a `CheckFunc`, [you may notice that it is an empty `interface{}`](../pkg/checker/checker.go)!

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

If you need to do something like this, you may want to define such a custom check using `checker.NewChainedCheckFunc`, which simplifies the declaration for such a function; you can put your custom struct in the function signature of the function passed into `checker.NewChainedCheckFunc`; the struct's type will be inferred to be the value of the type parameter `S`.

#### Dealing With Custom Resources

If you plan to write a `checker.ChainedCheckFunc` on a custom resource in Kubernetes, the major caveat is that you need to ensure that any Kubernetes resource Go types are added to the `pkg/checker.Scheme` defined [here](../pkg/checker/scheme.go).

For users who have developed Kubernetes controllers before, it may be familiar to mention that the usual way to do this is by running the `AddToScheme` function usually generated by most Go-based Kubernetes controller frameworks.

For example, here is how you would add the `corev1` Go types to ensure the Hull will be able to correctly identify that a YAML object that has `apiVersion: v1` and `kind: ConfigMap` should be marshalled into the `*corev1.ConfigMap` struct:

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

If you also need your `CheckFunc` to be able to work with other workload types, you may also need to import `appsv1` and `batchv1`, so you can modify this accordingly:

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

> **Note**: Why is this necessary?
>
> Without adding a type to the underlying `checker.Scheme`, Hull will not understand how to convert a YAML object it sees with a given `apiVersion` and `kind` to a corresponding Go struct that is within your function signature.
>
> If a type cannot be found in the `checker.Scheme`, it will **always** be assumed that your object is of type [`*unstructured.Unstructured`](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured#Unstructured).
>
> You can also use `*unstructured.Unstructured` as a catch-all for objects of all types in `checker.CheckFunc`s; for example, if you wanted to check that all objects, regardless of type, have the expected labels, you will want to use this type.

#### Adding a real `test.NamedCheck`

Now that we've gone over what `test.Checks` are, we can fix our previous suite to have a real check contained in our `suite.NamedChecks`. Let's also rename the check accordingly:

```go
var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{
					"hello": "cattle",
				}),
		},
	},

	NamedChecks: []test.NamedCheck{
		{
			Name: "ConfigMaps have expected data",

			Covers: []string{".Values.data"},

			Checks: test.Checks{
				checker.PerResource(func(tc *checker.TestContext, configmap *corev1.ConfigMap) {
					assert.Contains(tc.T,
						configmap.Data, "config",
						"%T %s does not have 'config' key", configmap, checker.Key(configmap),
					)
					if tc.T.Failed() {
						return
					}
					assert.Equal(tc.T,
						checker.ToYAML(checker.MustRenderValue[map[string]string](tc, ".Values.data")), configmap.Data["config"],
						"%T %s does not have correct data in 'config' key", configmap, checker.Key(configmap),
					)
				}),
			},
		},
	},
}
```

If we run this, we should find that the `simple-chart` passes this test! Now we're ready to move onto our next check.

> **Note**: `checker.MustRenderValue` is a generic function, so it can return any type that you expect would belong in the
> path provided. 
>
> Here, we expect a `map[string]string`, but it would be valid to provide a struct here, or even a pointer to a struct.
>
> The only caveat though is that `checker.MustRenderValue` **cannot return `nil`**. Therefore, in cases where the value can be `nil`, you should use `val, ok := checker.RenderValue(...)`, where `!ok` signifies that either nothing was found in that path (it was never set) or the value at that path is set to `nil`.
>
> Here, I'm assuming it's safe to use `.MustRenderValue` since the `simple-chart` always provides `.Values.data` as a default value, but a user setting `.Values.data` in the `values.yaml` to `null` can cause this check to **panic (not fail)** due to the use of `MustRenderValue` instead of `RenderValue`. You can try this out by adding a `test.Case` that does this!

> **Note**: How did I know to `import corev1 "k8s.io/api/core/v1"` to get the type for a `*corev1.ConfigMap`?
>
> For most built-in Kubernetes resources, the definition for the type can always be found at `k8s.io/api/GROUP/VERSION`; since `ConfigMaps` is in the default group (`""`, which translates to `core`) and the version of the group I'm looking for is `v1`, I found it at `k8s.io/api/core/v1`.
>
> Similarly, a `ClusterRole` belongs to the API group `rbac` at version `v1`, so `import rbacv1 "k8s.io/api/rbac/v1"` would allow me to use `*rbac.ClusterRole`.
>
> Occasionally, when defining fields like `metadata.*` in the resources, you may also need to import `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`; this library contains the structs that are commonly used by all Kubernetes resources.
>
> Remember to always use the pointer (`*corev1.ConfigMap`) not the value (`corev1.ConfigMap`) since the pointer is the one that implements the interface `runtime.Object` and `metav1.Object`, not the value!

#### Adding a `test.FailureCase`

Sometimes, you want to test conditions where you **expect** a template failure, such as when you have `fail` or `required` block in your templates; in these cases, you don't need to run any checks, but you do want to ensure that a user gets the right error.

To do this, you can add a `test.FailureCase`; it's fairly similar to a `test.Case`, but with the ability to declare coverage (since there is no produced manifest to run `suite.NamedCheck`s on) and the ability to assert that a specific failure message is received.

Let's finish up our coverage by adding the last two `test.FailureCase`s, covering our full `values.yaml`:

```go
var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name:            "Using Defaults",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
		{
			Name: "Override .Values.data",
			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				Set("data", map[string]string{
					"hello": "cattle",
				}),
		},
	},

	NamedChecks: []test.NamedCheck{
		{
			Name:   "Test",
			Covers: []string{".Values.data"},
			Checks: test.Checks{
				checker.PerResource(func(tc *checker.TestContext, configmap *corev1.ConfigMap) {
					assert.Contains(tc.T,
						configmap.Data, "config",
						"%T %s does not have 'config' key", configmap, checker.Key(configmap),
					)
					if tc.T.Failed() {
						return
					}
					assert.Equal(tc.T,
						checker.ToYAML(checker.MustRenderValue[map[string]string](tc, ".Values.data")), configmap.Data["config"],
						"%T %s does not have correct data in 'config' key", configmap, checker.Key(configmap),
					)
				}),
			},
		},
	},

	FailureCases: []test.FailureCase{
		{
			Name: "Set .Values.shouldFail",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("shouldFail", "true"),

			Covers: []string{
				".Values.shouldFail",
			},

			FailureMessage: ".Values.shouldFail is set to true",
		},
		{
			Name: "Set .Values.shouldFailRequired",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace).
				SetValue("shouldFailRequired", "true"),

			Covers: []string{
				".Values.shouldFailRequired",
			},

			FailureMessage: ".Values.shouldFailRequired is set to true",
		},
	},
}
```

Coverage should now be passing for the full chart! 

You can see the full working example at [`../examples/tests/simple`](../examples/tests/simple/) or a more complex example at [`../examples/tests/example`](../examples/tests/example/) that does not currently have full coverage (this is left as an exercise to the reader).

## Should I Use Hull?

### Is this the best Helm testing framework for me?

Depends on your use case!

Hull is great for organization that:
- Needs a fairly robust Helm template testing tool with easy extensibility
- Works primarily in Golang (such as those that develop Golang Kubernetes operators and ship them in Helm charts)

For example, Hull is targeted for use in [`rancher/charts`](https://github.com/rancher/charts/tree/main/charts) today, a repository that maintains a large number of highly complex charts that primarily deploy operators built in Go.
