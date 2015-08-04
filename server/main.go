package main

import (
	"log"
	"mutantfortune/ifc"
	"mutantfortune/server/util"
	"mutantfortune/service"
	"v.io/v23"
	"v.io/x/ref/lib/signals"
	_ "v.io/x/ref/runtime/factories/generic"
)

var (
	theServiceName = "croupier"
)

func main() {
	ctx, shutdown := v23.Init()
	defer shutdown()

	// v23.GetNamespace(ctx).SetRoots("/monopole2.mtv.corp.google.com:23000")
	v23.GetNamespace(ctx).SetRoots("/104.197.96.113:3389")
	// A generic server.
	s := util.MakeServer(ctx)

	// Attach the 'fortune service' implementation
	// defined above to a queriable, textual description
	// of the implementation used for service discovery.
	fortune := ifc.FortuneServer(service.Make())

	// If the dispatcher isn't nil, it's presumed to have
	// obtained its authorizer from util.MakeAuthorizer().
	dispatcher := util.MakeDispatcher()

	// Start serving.
	var err error
	if dispatcher == nil {
		// Use the default dispatcher.
		err = s.Serve(
			theServiceName, fortune, util.MakeAuthorizer())
	} else {
		err = s.ServeDispatcher(theServiceName, dispatcher)
	}
	if err != nil {
		log.Panic("Error serving service: ", err)
	}
	<-signals.ShutdownOnSignals(ctx)
}
