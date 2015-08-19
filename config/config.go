//  See https://github.com/golang/mobile/blob/master/example/network/main.go
//
// # disconnect phone
// # turn off developer options
// # turn it back on, be sure that USB debugging is enabled
// adb kill-server
// adb start-server
// # plug the device in
// adb devices
// adb logcat | grep GoLog
//
// export GITDIR=github.com/monopole/croupier
//
// $BERRY/bin/mounttabled --v23.tcp.address :23000 &
//
// export MT_HOST=192.168.43.136
//
// $BERRY/bin/namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
//
// GOPATH=$BERRY $BERRY/bin/gomobile install $GITDIR/volley

package config

const Chatty = true

const RootName = "volley/player"

// const MountTableHost = "104.197.96.113" // Asim's gce instance
// const MountTableHost = "192.168.2.71"  // my laptop on house net
// const MountTableHost = "192.168.43.136" // my latop on motox net
const MountTableHost = "localhost"

// const MountTablePort = "3389" // Asim's preferred port
const MountTablePort = "23000"

const NamespaceRoot = "/" + MountTableHost + ":" + MountTablePort

const FailFast = true
const TestDomain = "http://www.example.com"
