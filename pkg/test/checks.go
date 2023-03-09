package test

import "github.com/aiyengar2/hull/pkg/checker"

// NamedCheck is a check that run on every `helm template` call identified by the test Suite
type NamedCheck struct {
	Name   string
	Checks Checks
	Covers []string
}

type Checks []checker.ChainedCheckFunc
