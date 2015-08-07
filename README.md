# croupier
Multi-device Go/Gl/v23 demo

Steps to get v23 running in Go directly on mobile, and reproduce v23 bugs.

The procedure below writes to `~/pumpkin`; adjust as desired.

```
unset GOROOT
unset GOPATH
export PUMPKIN=~/pumpkin
```

## Install android-sdk-linux

You'll need `adb` on your `PATH`.  See the
[instructions](https://developer.android.com/sdk/index.html).

## Install go 1.5 beta

You'll need bleeding edge 1.5 `go` on your `PATH`.  See the
[instructions](http://golang.org/doc/install/source).

## Clear your environment

```
/bin/rm -rf $PUMPKIN
mkdir $PUMPKIN
```

## Install v23 as an end-user

Installing as an end-user is easier than installing as a
[contributor](https://v.io/community/contributing.html).

__Because of code mirror server failures, one may have to repeat these
incantations a few times.__

```
GOPATH=$PUMPKIN go get -d v.io/x/ref/...
GOPATH=$PUMPKIN go install v.io/x/ref/cmd/...
GOPATH=$PUMPKIN go install v.io/x/ref/services/agent/...
GOPATH=$PUMPKIN go install v.io/x/ref/services/mounttable/...
```

Otherwise follow the full
[instructions](https://v.io/installation/details.html).

## Install Go mobile stuff

```
GOPATH=$PUMPKIN go get golang.org/x/mobile/cmd/gomobile
GOPATH=$PUMPKIN $PUMPKIN/bin/gomobile init
```

## Install croupier

Create and fill `$PUMPKIN/src/github.com/monopole/croupier'.

It may complain about _No buildable Go source_ - no worries.

```
gitmf=github.com/monopole/croupier
GOPATH=$PUMPKIN go get -d ${gitmf}
```

Now generate the Go that was missing and build the v23 fortune server
and client stuff.

```
VDLROOT=$PUMPKIN/src/v.io/v23/vdlroot \
    VDLPATH=$PUMPKIN/src \
    $PUMPKIN/bin/vdl generate --lang go $V_TUT/src/${gitmf}/ifc

GOPATH=$PUMPKIN go build ${gitmf}/ifc
GOPATH=$PUMPKIN go build ${gitmf}/service
GOPATH=$PUMPKIN go install ${gitmf}/client
GOPATH=$PUMPKIN go install ${gitmf}/server
```

## Test desktop mode

Build `croupier` for the  desktop.
This app is a small modification of the
[gomobile basic example](https://godoc.org/golang.org/x/mobile/example/basic).

```
GOPATH=$PUMPKIN go install ${gitmf}/croupier
```

Check the namespace, make sure there's nothing that looks like `croupier*`
```
$V_BIN/namespace --v23.namespace.root '/104.197.96.113:3389' glob  'croupier*'
```

Open another terminal and run
```
$PUMPKIN/bin/croupier 
```

You should see a new window with a triangle.

Open yet _another_ terminal and run
```
$PUMPKIN/bin/croupier 
```
This window should not have a triangle.

Drag the triangle in the first window.
On release, it should hop to the second window.
It should be possible to send it back.

The `namespace` command above should now show two services, `croupier0` and `croupier1`

__To run with more than two devices (a 'device' == a desktop terminal
or an app running on a phone), one must change the the constant
`expectedInstances` in the file
[game_manager.go](https://github.com/monopole/mutantfortune/blob/master/croupier/util/game_manager.go).__


## Now try the mobile version

The mobile app counts as a "device" against the  limit set by
`expectedInstances`, so for the default value of two, only
one desktop terminal is allowed.

Plug your dev phone into a USB port.

Enter this:

```
GOPATH=$PUMPKIN:~/pumpkin $V_BIN/gomobile install ${gitmf}/croupier
```

You should see a triangle (or not) depending on the order in which you launched it with
respect to other instances of the app.

To debug:

```
adb logcat > log.txt
```
