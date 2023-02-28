package test

import "github.com/aiyengar2/hull/pkg/checker"

// TemplateCheck is a check that run on every `helm template` call identified by the test Suite
type TemplateCheck struct {
	Name string
	Func checker.CheckFunc

	// OmitCases contains a list of names of test.Cases that, if overridden, would cause a DefaultCheck to skip executing itself on that check
	OmitCases []string
}

// ValueCheck is a check that is run on a specific `helm template` call with a given set of values.yaml identified by a given test Case
type ValueCheck struct {
	Name string
	Func checker.CheckFunc

	// Covers is a list of file globs whose objects are covered by this template check
	// This is used for coverage to be able to identify if the logic in certain files has been checked
	Covers []string
}
