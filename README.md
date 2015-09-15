# volley

Go + mobile + GL + [v23](https://v.io).
[Demo video](https://youtu.be/dk-5BkO_P14).
To try v23, sign up [here](https://goo.gl/ETo8Mt).

## Install prerequisites

Define a clean env to make the rest of the procedures described here
more likely to work.

```
unset GOROOT
unset GOPATH
originalPath=$PATH
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```


### X11 development (ubuntu)

Install supplemental GL libs
See notes [here](https://github.com/golang/mobile/blob/master/app/x11.go#L15).
```
sudo apt-get install libegl1-mesa-dev libgles2-mesa-dev libx11-dev
```


### Android development

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

### OSX/iOS development


#### Become an app developer 

* Become an [Apple app developer](https://developer.apple.com/programs) (get an apple ID, device auth, etc.)
* Install XCode, perhaps: `xcode-select --install`
* Get [git](http://git-scm.com/download/mac).

#### install [ios-deploy](https://github.com/phonegap/ios-deploy)

```
brewUrl=https://raw.githubusercontent.com/Homebrew/install/master/install
ruby -e "$(curl -fsSL $brewUrl)"
brew install node
npm install -g ios-deploy
make install prefix=/usr/local
ios-deploy
```

### Install Go 1.5


#### Get binaries

Full instructions [here](https://golang.org/doc/install).
On a 64-bit linux box, just try this:

```
sudo /bin/rm -rf /usr/local/go_old
sudo mv -n /usr/local/go /usr/local/go_old
( tarball=https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz
  curl $tarball -o - | sudo tar -C /usr/local -xzf - )
```

#### Add to PATH
```
PATH=/usr/local/go/bin:$PATH
go version
```


## Define workspace

The remaining commands destructively write to the directory
pointed to by `BERRY`.

```
export BERRY=~/pumpkin
```

Optionally wipe it
```
/bin/rm -rf $BERRY
mkdir $BERRY
```

Modify paths to make later commands shorter.
```
PATH=$BERRY/bin:$PATH
export GOPATH=$BERRY
```

## Install v23 as an end-user

To try v23, sign up [here](https://goo.gl/ETo8Mt).

Full instructions [here](https://v.io/installation/details.html), or try this:

__Because of code mirror server failures, one may have to repeat these
incantations a few times.__

```
go get -d v.io/x/ref/...
go install v.io/x/ref/cmd/...
go install v.io/x/ref/services/agent/...
go install v.io/x/ref/services/mounttable/...
```

## Install Go mobile stuff

```
go get golang.org/x/mobile/cmd/gomobile
gomobile init
```

## Install volley software

Create and fill `$BERRY/src/github.com/monopole/volley`.

_Ignore_ complaints about `No buildable Go source`.

```
GITDIR=github.com/monopole/volley
```

Grab the code:
```
go get -d $GITDIR
```

Generate the Go that was missing and build the v23 fortune server
and client stuff.

```
VDLROOT=$BERRY/src/v.io/v23/vdlroot \
    VDLPATH=$BERRY/src \
    $BERRY/bin/vdl generate --lang go $BERRY/src/$GITDIR/ifc
```

## Setup your network

Each game instance must be able to communicate with other instances
and with a v23 mounttable (discussed below).

One means to allow this is to start a wifi access point on a phone and
connect everything to it.

### __Drop firewalls__


In the exercise below, one must run a mounttable.  On the computer
involved, or any other computers, forward ports appropriately or
simply drop the firewall.

E.g. on linux try this sledgehammer
```
sudo $BERRY/src/github.com/monopole/volley/dropFw.sh 
```
On Mac, something like
```
 Apple Menu -> System Preferences -> Security & Privacy -> Firewall tab -> disable
```

__Don't run any game instances until this is done__,
as they may hang on network attempts without any feedback (bug).


### Run a mounttable

Pick a computer on the network and discover its IP address.

If you just want to try volley on a local workstation,
do this:
```
export MT_HOST=127.0.0.1
```

If you want to run a demo with multiple machines, or mobile devices,
using a mounttable you control, it's simplest to use local net
addresses (i.e. `192.168.*.*`) assigned by, say, your WAP over DHCP.

Store this important address in an env var:

```
sedPa="s/.*addr:(192\.168\.[0-9]+\.[0-9]+).*/\1/p"
export MT_HOST=`ifconfig | grep "inet addr:192.168" | sed -rn $sedPa`
```

Then make a full V23 root spec thus:
```
MT_PORT=23000

V23_NS_ROOT="$MT_HOST:$MT_PORT"

echo V23_NS_ROOT=$V23_NS_ROOT
```

On the computer that `$MT_HOST` points to, run a mounttable daemon.
Game instances need this to find each other.

```
mounttabled --v23.tcp.address $V23_NS_ROOT &
```

To verify that the table is up and the network is WAI,
try the following on another local network computer (after
setting `V23_NS_ROOT` to the appropriate value):

```
namespace --v23.namespace.root /$V23_NS_ROOT glob -l '*/*'
```

This is basically `ls` for the mount table.

If this __immediately__ returns with no output, that's good.  It means
the mounttable was contacted and is empty.  Later, when games are
running, all game instances will appear in the table.

If the request appears to hang, eventually timing out, then something
is wrong with the network.  Try pinging.  Try shutting down firewalls.

### Edit the app config

__TODO(monopole):  This section out of date.  `volley` now pulls the MT address from [trustybike.net](http://trustybike.net).
Must convert to neighborhood (or something)__

The _discovery_ aspect of the game hasn't had any work done yet,
so one must hardcode the mounttable's `IP:port` in the app
before building and deploying it. 

Do so by changing the value of `MountTableHost` in
[`volley/config/config.go`](https://github.com/monopole/volley/blob/master/config/config.go)
to the value of `$MT_HOST`.

Change the port too if necessary.

If the current value is `127.0.0.1`, these linux commands do the job:
```
(cf=$BERRY/src/$GITDIR/config/config.go
 sed -i s/127.0.0.1/$MT_HOST/ $cf
 grep MountTableHost $cf)
```

## Build and Run

Build `volley` for the desktop.

```
go install $GITDIR/volley
```

Check the namespace:
```
namespace --v23.namespace.root /$V23_NS_ROOT glob -l '*/*'
```

Open another terminal and run (adjusting path as needed obviously)
```
~/pumpkin/bin/volley
```

You should see a new window with a triangle.

Open yet _another_ terminal and run
```
~/pumpkin/bin/volley
```

Quickly swipe the triangle in the first window.  It should
hop to the second window.

The `namespace` command above should now show two entries:
`volley/player0001` and `volley/player0002`

## Try the mobile device version

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
gomobile install $GITDIR/volley
```

The app should appear as `aaaVolley` per the `application` tag
in the [manifest](https://github.com/monopole/volley/blob/master/volley/AndroidManifest.xml).

Just click to run, there's no further setup.

__Be sure your device is on the local network, so it has a chance of
seeing the mounttable, or it simply won't work.__

It the app finds the mounttable, it will join the "game"
(or start a new one).

### ioS

```
gomobile init
$BERRY/bin/gomobile build -target=ios $GITDIR/volley
ios-deploy --bundle volley.app
```

__Be sure your device is on the local network, so it has a chance of
seeing the mounttable, or it simply won't work.__

It the app finds the mounttable, it will join the "game"
(or start a new one).
