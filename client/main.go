package main

import (
	"flag"
	"fmt"
	"time"

	"mutantfortune/ifc"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/options"
	"v.io/x/lib/vlog"
	_ "v.io/x/ref/runtime/factories/generic"
)

var (
	theServiceName = "croupier"
	server         = flag.String(
		"server", "", "Name of the server to connect to")
	newFortune = flag.String(
		"add", "", "A new fortune to add to the server's set")
)

func main() {
	ctx, shutdown := v23.Init()
	defer shutdown()

	v23.GetNamespace(ctx).SetRoots("/104.197.96.113:3389")
	//v23.GetNamespace(ctx).SetRoots("/monopole2.mtv.corp.google.com:23000")

	//	if *server == "" {
	//		vlog.Error("--server must be specified")
	//		return
	//	}
	//	f := ifc.FortuneClient(*server)
	f := ifc.FortuneClient(theServiceName)
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if *newFortune == "" { // --add flag not specified
		fortune, err := f.Get(ctx, options.SkipServerEndpointAuthorization{})
		if err != nil {
			vlog.Errorf("error getting fortune: %v", err)
			return
		}
		fmt.Println(fortune)
	} else {
		if err := f.Add(ctx, *newFortune, options.SkipServerEndpointAuthorization{}); err != nil {
			vlog.Errorf("error adding fortune: %v", err)
			return
		}
	}
}
