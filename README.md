# croupier
Multi-device Go+GL+v23 demo.


## Install prerequisites

For bootstrapping, prefer a very clean environment.

```
unset GOROOT
unset GOPATH
originalPath=$PATH
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```

### Set up Android development

Android's `adb` is a prerequisite.

#### Install

Full instructions
[here](https://developer.android.com/sdk/index.html), or try this:
```
cd
/bin/rm -rf android-sdk-linux
curl http://dl.google.com/android/android-sdk_r24.3.3-linux.tgz -o - | tar xzf -
cd android-sdk-linux

# Answer ‘y’ a bunch of times, after consulting with an attorney.
./tools/android update sdk --no-ui

# Might help to do this:
sudo apt-get install lib32stdc++6
```

#### Add to path

```
PATH=~/android-sdk-linux/platform-tools:$PATH
~/android-sdk-linux/platform-tools/adb version
adb version
```

### Setup for iOS development

_TBD_

### Install go 1.5 beta

Go 1.5 required (still beta as of July 2015).

#### Install

Full instructions [here](http://golang.org/doc/install/source), or try this:

```
cd

# The following writes to ./go
/bin/rm -rf go
curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz -o - \
    | tar xzf -

# Get this ‘go’ out of the way for go1.5 beta (which needs 1.4.2 to build it)
mv go go1.4.2

# Build Go from head per http://golang.org/doc/install/source
git clone https://go.googlesource.com/go
cd go
git checkout master
cd src
GOROOT_BOOTSTRAP=$HOME/go1.4.2 ./all.bash
```


#### Add to PATH
```
PATH=~/go/bin:$PATH
~/go/bin/go version
go version
```


## Define workspace

The procedure below writes to `~/pumpkin` (arbitary)
Adjust as desired.

```
export PUMPKIN=~/pumpkin
```

Optionally wipe it
```
/bin/rm -rf $PUMPKIN
mkdir $PUMPKIN
```


## Install v23 as an end-user

Full instructions [here](https://v.io/installation/details.html), or try this:

__Because of code mirror server failures, one may have to repeat these
incantations a few times.__

```
GOPATH=$PUMPKIN go get -d v.io/x/ref/...
GOPATH=$PUMPKIN go install v.io/x/ref/cmd/...
GOPATH=$PUMPKIN go install v.io/x/ref/services/agent/...
GOPATH=$PUMPKIN go install v.io/x/ref/services/mounttable/...
```

## Install Go mobile stuff

```
GOPATH=$PUMPKIN go get golang.org/x/mobile/cmd/gomobile
GOPATH=$PUMPKIN $PUMPKIN/bin/gomobile init
```

## Install croupier

Create and fill `$PUMPKIN/src/github.com/monopole/croupier`.

Ignore complaints about _No buildable Go source_.

```
gitdir=github.com/monopole/croupier
```

Grab the code:
```
GOPATH=$PUMPKIN go get -d ${gitdir}
```

Generate the Go that was missing and build the v23 fortune server
and client stuff.

```
VDLROOT=$PUMPKIN/src/v.io/v23/vdlroot \
    VDLPATH=$PUMPKIN/src \
    $PUMPKIN/bin/vdl generate --lang go $PUMPKIN/src/${gitdir}/ifc

GOPATH=$PUMPKIN go build ${gitdir}/ifc
GOPATH=$PUMPKIN go build ${gitdir}/service
GOPATH=$PUMPKIN go install ${gitdir}/client
GOPATH=$PUMPKIN go install ${gitdir}/server
```

## Test desktop mode

Build `croupier` for the  desktop.
This app is a small modification of the
[gomobile basic example](https://godoc.org/golang.org/x/mobile/example/basic).

```
GOPATH=$PUMPKIN go install ${gitdir}/croupier
```

Check the namespace, make sure there's nothing that looks like `croupier*`
```
$PUMPKIN/bin/namespace --v23.namespace.root '/104.197.96.113:3389' glob  'croupier*'
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
GOPATH=$PUMPKIN:~/pumpkin $PUMPKIN/bin/gomobile install ${gitdir}/croupier
```

You should see a triangle (or not) depending on the order in which you launched it with
respect to other instances of the app.

To debug:

```
adb logcat > log.txt
```
