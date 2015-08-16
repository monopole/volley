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
// $VEGGIE/bin/mounttabled --v23.tcp.address :23000 &
//
// $VEGGIE/bin/namespace --v23.namespace.root '/localhost:23000' glob -l '*/*'
//
// $VEGGIE/bin/beans

package config

const Chatty = true

const RootName = "volley/player"

// const MountTableHost = "localhost"
// const MountTableHost = "104.197.96.113"
// const MountTableHost = "172.17.166.64"
// const MountTableHost = "192.168.2.71"

const MountTableHost = "192.168.43.136"
const MountTablePort = "23000"

const NamespaceRoot = "/" + MountTableHost + ":" + MountTablePort

const FailFast = true
const TestDomain = "http://www.example.com"
