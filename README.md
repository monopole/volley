# mutantfortune

Steps to get v23 running in Go directly on mobile, and reproduce v23 bugs.

The procedure below puts code in the two directories

```
 ~/pumpkin
 ~/mutantfortune
```

Feel free to change the names in the commands below.

The directories can be deleted afterwards.

Goal is to just hack something together without messing up existing
installs, and clean it up in a proper repo once things seem to be
on the right track.


## Install android-sdk-linux

See the [instructions](https://developer.android.com/sdk/index.html).

You'll need `adb`


## Install go 1.5 beta

See the [instructions](http://golang.org/doc/install/source).

## Install the go mobile stuff

Using `pumpkin` arbitrarily.  Use whatever you want.

```
unset GOROOT
unset GOPATH
/bin/rm -rf ~/pumpkin
mkdir ~/pumpkin

GOPATH=~/pumpkin go get golang.org/x/mobile/cmd/gomobile
GOPATH=~/pumpkin ~/pumpkin/bin/gomobile init
```

## Install v23 as an end-user

...as opposed to the more complex install for a contributor.


__Because of permission failures, you may have to repeat these incantations a few times.__


```
GOPATH=~/pumpkin go get -d v.io/x/ref/...
GOPATH=~/pumpkin go install v.io/x/ref/cmd/...
GOPATH=~/pumpkin go install v.io/x/ref/services/agent/...
GOPATH=~/pumpkin go install v.io/x/ref/services/mounttable/...
```

Otherwise follow the full
[instructions](https://v.io/installation/details.html).

## Install the fortune client and server

```
export V23_RELEASE=~/pumpkin
export V_BIN=$V23_RELEASE/bin
export V_TUT=~/mutantfortune

/bin/rm -rf ~/mutantfortune

gitmf=github.com/monopole/mutantfortune

GOPATH=~/mutantfortune go get -d ${gitmf}

VDLROOT=$V23_RELEASE/src/v.io/v23/vdlroot VDLPATH=$V_TUT/src \
  $V_BIN/vdl generate --lang go $V_TUT/src/${gitmf}/ifc

GOPATH=~/mutantfortune:~/pumpkin go build ${gitmf}/ifc

GOPATH=~/mutantfortune:~/pumpkin go install ${gitmf}/client

GOPATH=~/mutantfortune:~/pumpkin go install ${gitmf}/server
```


## Test the basic fortune app (no mobile involved)

Check the mount table to confirm there is NO service named `croupier`
```
$V_BIN/namespace --v23.namespace.root '/104.197.96.113:3389' glob -l '*'
```

Start the service - it is hardcoded to load itself into the public
mount table at 104.blah.blah:blah (from above)

```
$V_TUT/bin/server &
TUT_PID_SERVER=$!
```

Check again - this time the output should include `croupier`
```
$V_BIN/namespace --v23.namespace.root '/104.197.96.113:3389' glob -l '*'
```

Use the client to obtain a fortune.  It finds the service via the public mount table.
```
$V_TUT/bin/client
```

Kill the service, and confirm that `croupier` no longer in the table.
```
kill $TUT_PID_SERVER

$V_BIN/namespace --v23.namespace.root '/104.197.96.113:3389' glob -l '*'
```

## Now try the mobile app

Use this to install an app called `croupier`, which will NOT do networking

```
GOPATH=~/mutantfortune:~/pumpkin $V_BIN/gomobile install ${gitmf}/croupier
```

Run it to see the purple triangle - a trivial change from
https://godoc.org/golang.org/x/mobile/example/basic

Now edit the file and change `doGl := false` to true, reinstall, and
see the program fail.

Look for `GoLog` in the output of

```
adb logcat > log.txt
```


