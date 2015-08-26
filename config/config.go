// export MT_HOST=192.168.43.136
//
// $BERRY/bin/mounttabled --v23.tcp.address :23000 &
//
// $BERRY/bin/namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
//
// GOPATH=$BERRY go install $GITDIR/volley
//
// GOPATH=$BERRY $BERRY/bin/gomobile install $GITDIR/volley

package config

const (
	// MountTableHost = "104.197.96.113" // Asim's gce instance
	// MountTableHost = "192.168.2.71" // my laptop on home net
	MountTableHost = "192.168.43.136" // my laptop on motox net
	// MountTableHost = "127.0.0.1"

	// MountTablePort = "3389" // Asim's preferred port
	MountTablePort = "23000"
)

const (
	NamespaceRoot = "/" + MountTableHost + ":" + MountTablePort
	FailFast      = false
	TestDomain    = "http://www.example.com"
	MagicX        = -99
	Chatty        = false
	RootName      = "volley/player"
)
