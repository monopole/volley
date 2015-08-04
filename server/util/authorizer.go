package util

import (
	"flag"
	"v.io/v23/context"
	"v.io/v23/security"
)

var (
	openStart = flag.Int(
		"start", 12, "Hour when friends may start access.")
	openLength = flag.Int(
		"length", 1, "Number of hours the window stays open.")
)

type policy struct{}

func (policy) Authorize(ctx *context.T, call security.Call) error {
	return nil
}

func MakeAuthorizer() security.Authorizer {
	return policy{}
}
