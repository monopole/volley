# volley
An exploration of Go + GL + mobile + v23

## Install prerequisites

For bootstrapping, prefer a clean environment.

```
unset GOROOT
unset GOPATH
originalPath=$PATH
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```


### Setup for X11 development (ubuntu)

Install supplemental GL libs.
See notes [here](https://github.com/golang/mobile/blob/master/app/x11.go#L15).
```
sudo apt-get install libegl1-mesa-dev libgles2-mesa-dev libx11-dev
```


### Setup for Android development

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


#### Become an app developer 


* Install XCode, perhaps: `xcode-select --install`
* Get [git](http://git-scm.com/download/mac).
* Get provisioned to become an ios app developer (apple ID, etc.)

#### install [ios-deploy](https://github.com/phonegap/ios-deploy)

Maybe:
```
ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
brew install node
npm install -g ios-deploy
make install prefix=/usr/local
ios-deploy
```



### Install Go 1.5 beta

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
# GOROOT_BOOTSTRAP=/usr/local/go ./all.bash
GOROOT_BOOTSTRAP=$HOME/go1.4.2 ./all.bash
```


#### Add to PATH
```
PATH=~/go/bin:$PATH
~/go/bin/go version
go version
```


## Define workspace

The remaining commands destructively write to the directory
pointed to by `BERRY`.

```
export BERRY=~/pumpkin
PATH=$BERRY/bin:$PATH
```

Optionally wipe it
```
/bin/rm -rf $BERRY
mkdir $BERRY
```


## Install v23 as an end-user

Full instructions [here](https://v.io/installation/details.html), or try this:

__Because of code mirror server failures, one may have to repeat these
incantations a few times.__

```
GOPATH=$BERRY go get -d v.io/x/ref/...
GOPATH=$BERRY go install v.io/x/ref/cmd/...
GOPATH=$BERRY go install v.io/x/ref/services/agent/...
GOPATH=$BERRY go install v.io/x/ref/services/mounttable/...
```

Alternatively, try this [tarball](https://drive.google.com/a/google.com/file/d/0B_KAJdV1hyzkQndMLTBabUcxdGM/view).

## Fix crypto libs

See https://vanadium.googlesource.com/third_party/+log/master/go/src/golang.org/x/crypto

Edit these two files, _adding_ a bogus hardware (e.g. `,jeep`) target to
the `+build` line:

```
$BERRY/src/golang.org/x/crypto/poly1305/poly1305_arm.s
$BERRY/src/golang.org/x/crypto/poly1305/sum_arm.go
```

Edit this file, *removing* `,!arm` from the `+build` line:

```
$BERRY/src/golang.org/x/crypto/poly1305/sum_ref.go
```


## Install Go mobile stuff

```
GOPATH=$BERRY go get golang.org/x/mobile/cmd/gomobile
GOPATH=$BERRY gomobile init
```

## Install volley software

Create and fill `$BERRY/src/github.com/monopole/croupier`.

Ignore complaints about _No buildable Go source_.

```
GITDIR=github.com/monopole/croupier
```

Grab the code:
```
GOPATH=$BERRY go get -d $GITDIR
```

Generate the Go that was missing and build the v23 fortune server
and client stuff.

```
VDLROOT=$BERRY/src/v.io/v23/vdlroot \
    VDLPATH=$BERRY/src \
    $BERRY/bin/vdl generate --lang go $BERRY/src/$GITDIR/ifc
```

## Setup your network

__All devices that are part of the game need to be able to find a v23
`mounttable` and each other.__


### Get on a local network, e.g.

* Open a wifi access point on a phone.
* Connect all devices to it - they should
  get ip numbers like `192.168.*.*`
* __Drop firewall__ on your laptop, e.g. on linux
  try [this script](https://github.com/monopole/croupier/blob/master/nofw.sh).

__Don't run any game instances until this is done__, as the game may
hang on network attempts without any feedback.

### Run a mounttable

Pick a laptop on the network and discover its IP address.

```
ifconfig | grep "inet addr"
```

Store this important address in an env var (replacing __x__ appropriately):
```
export MT_HOST=192.168.x.x
```

On said laptop, run a mounttable.
Game instances need this to find each other.

```
$BERRY/bin/mounttabled --v23.tcp.address /${MT_HOST}:23000 &
```

To verify that the table is up, on another laptop connected to
the same WAP, run this

```
$BERRY/bin/namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
```

It should __immediately__ return with no output, indicating an empty mount
table.  If it's not fast, try pinging to confirm that the network is up.

Later, when games are running, all game instances will appear
in the table.


### Edit the app config.

Discovery is pretty bad right now.

One must hardcode the IP of the mounttable in the app
before building and deploying it.

Edit [`config.go`](https://github.com/monopole/croupier/blob/master/config/config.go#L32)
and change the line `MountTableHost` to
refer to the IP discussed above.


## Build and Run

Build `volley` for the  desktop.

This app derived from the
[gomobile basic example](https://github.com/golang/mobile/blob/master/example/basic/main.go)
([godoc here](https://godoc.org/golang.org/x/mobile/example/basic)).

```
GOPATH=$BERRY go install $GITDIR/volley
```

Check the namespace:
```
namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
```

Open another terminal and run
```
volley
```

You should see a new window with a triangle.

Open yet _another_ terminal and run
```
volley
```

Quickly swipe the triangle in the first window.  It should
hop to the second window.

The `namespace` command above should now show two entries:
`volley/player0001` and `volley/player0002`

## Try the mobile device version

__Be sure your device is on the local network, so it has a chance of
seeing the mounttable, or it simply won't work.__

Plug your device into a USB port.

### Android

Be sure you see the device from your computer.
A sequence that sometimes works is:

```
# Disconnect device
# In device, turn off developer options
# In device, turn it back on, be sure that USB debugging is enabled
adb kill-server
adb start-server
# Plug in the device.
adb devices
# The following reports log lines labelled GoLog.
adb logcat | grep GoLog
```

To deploy the app:
```
GOPATH=$BERRY gomobile install $GITDIR/volley
```

The app should appear as `aaaVolley` per the `application` tag
in the [manifest](https://github.com/monopole/croupier/blob/master/volley/AndroidManifest.xml).

Just click to run, there's no further setup.

It the app finds the mounttable, it will join the game.

### ioS

```
GOPATH=$BERRY gomobile init
GOPATH=$BERRY $BERRY/bin/gomobile build -target=ios $GITDIR/volley
ios-deploy --bundle volley.app
```

Just click to run, there's no further setup.

It the app finds the mounttable, it will join the game.
