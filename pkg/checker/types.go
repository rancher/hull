package checker

type Check struct {
	Name    string
	Options *Options
	Func    interface{}
}

type Options struct {
	PerTemplateManifest bool
}

type CheckFunc interface{}
