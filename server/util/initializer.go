package util

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"

	"v.io/v23/rpc"
)

var (
	fileName = flag.String(
		"endpoint-file-name", "",
		"Write endpoint address to given file.")
)

func saveEndpointToFile(e naming.Endpoint) {
	contents := []byte(
		naming.JoinAddressName(e.String(), "") + "\n")
	if ioutil.WriteFile(*fileName, contents, 0644) != nil {
		log.Panic("Error writing ", *fileName)
	}
	fmt.Printf("Wrote endpoint name to %v.\n", *fileName)
}

func MakeServer(ctx *context.T) rpc.Server {
	//	s, err := v23.NewServer(ctx, options.SecurityNone)
	s, err := v23.NewServer(ctx)
	if err != nil {
		log.Panic("Failure creating server: ", err)
	}

	endpoints, err := s.Listen(v23.GetListenSpec(ctx))
	if err != nil {
		log.Panic("Error listening to service: ", err)
	}
	if *fileName != "" {
		saveEndpointToFile(endpoints[0])
	}
	fmt.Printf("Listening at: %v\n", endpoints[0])
	return s
}
