package net

import (
	"flag"
	"io/ioutil"
	"log"

	"v.io/v23/rpc"
)

var (
	fileName = flag.String(
		"endpoint-file-name", "",
		"Write endpoint address to given file.")
)

func saveEndpointToFile(server rpc.Server) {
	endpoints := server.Status().Endpoints
	if len(endpoints) == 0 {
		return
	}
	for i, ep := range endpoints {
		log.Printf("Listening at endpoint %d of %d: %v", i+1, len(endpoints), ep)
	}
	if len(*fileName) == 0 {
		return
	}
	contents := []byte(endpoints[0].Name() + "\n")
	if ioutil.WriteFile(*fileName, contents, 0644) != nil {
		log.Panic("Error writing ", *fileName)
	}
	log.Printf("Wrote endpoint name to %v.\n", *fileName)
}
