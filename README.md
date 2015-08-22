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


### X11 development (ubuntu)

Install supplemental GL libs.
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



### Install Go 1.5


#### Install

Full instructions [here](https://golang.org/doc/install).
On a 64-bit linux box, just try this:

```
sudo mv -n /usr/local/go /usr/local/old_go
tarball=https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz
curl $tarball -o - | sudo tar -C /usr/local -xzf -
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

Full instructions [here](https://v.io/installation/details.html), or try this:

__Because of code mirror server failures, one may have to repeat these
incantations a few times.__

```
go get -d v.io/x/ref/...
go install v.io/x/ref/cmd/...
go install v.io/x/ref/services/agent/...
go install v.io/x/ref/services/mounttable/...
```

Alternatively, try this [tarball](https://drive.google.com/a/google.com/file/d/0B_KAJdV1hyzkQndMLTBabUcxdGM/view).


## Install Go mobile stuff

```
go get golang.org/x/mobile/cmd/gomobile
gomobile init
```

## Install volley software

Create and fill `$BERRY/src/github.com/monopole/croupier`.

_Ignore_ complaints about `No buildable Go source`.

```
GITDIR=github.com/monopole/croupier
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

## Setup your networks

__All devices that are part of the game need to be able to find a v23
`mounttable` and each other.__

This can be a challenge in a controlled corporate network.

A workaround is to open a wifi access point on a phone and
connect everything to it.

The devices will be given IP addresses `192.168.*.*`.

##### __Drop laptop firewalls__ 

In the exercise below, at least one laptop is required to run a
visible mounttable (not to mention build software etc.).

On that latop, or any other laptops which will run the game, drop the
firewall.

E.g. on linux try this port-unspecific hammer
```
sudo $BERRY/src/github.com/monopole/croupier/nofw.sh 
```

__Don't run any game instances until this is done__, as the game may
hang on network attempts without any feedback.


### Run a mounttable

Pick a laptop on the network and discover its IP address.

```
ifconfig | grep "inet addr"
```

Store this important address in an env var (replacing __x__ appropriately):
```
export MT_HOST=127.0.0.1  # For single machine testing obviously.
export MT_HOST=192.168.x.x
```

On said laptop, run a mounttable daemon.
Game instances need this to find each other.

```
mounttabled --v23.tcp.address ${MT_HOST}:23000 &
```

To verify that the table is up, on another laptop connected to
the same WAP, query it from another computer:

```
namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
```

This is basically `ls` for the mount table.

If this __immediately__ returns with no output, you're good - the
mounttable was contacted and is empty.  Later, when games are running,
all game instances will appear in the table.

If the request appears to hang, eventually timing out, then something
is wrong with the network.  Try pinging.  Try shutting down firewalls.

### Edit the app config.

The _discovery_ aspect of the game hasn't had any work done yet,
so one must hardcode the mounttable's `IP:port` in the app
before building and deploying it. 

Do so by changing the value of `MountTableHost` in
[`croupier/config/config.go`](https://github.com/monopole/croupier/blob/master/config/config.go)
to the value of `$MT_HOST`.

Change the port too if necessary.

If the current value is `127.0.0.1`, these linux commands do the job:
```
(cf=$BERRY/src/$GITDIR/config/config.go
 sed -i s/127.0.0.1/$MT_HOST/ $cf
 grep MountTableHost $cf)
```

## Build and Run

Build `volley` for the  desktop.

This app derived from the
[gomobile basic example](https://github.com/golang/mobile/blob/master/example/basic/main.go)
([godoc here](https://godoc.org/golang.org/x/mobile/example/basic)).

```
go install $GITDIR/volley
```

Check the namespace:
```
namespace --v23.namespace.root /${MT_HOST}:23000 glob -l '*/*'
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
gomobile install $GITDIR/volley
```

The app should appear as `aaaVolley` per the `application` tag
in the [manifest](https://github.com/monopole/croupier/blob/master/volley/AndroidManifest.xml).

Just click to run, there's no further setup.

It the app finds the mounttable, it will join the game.

### ioS

```
gomobile init
$BERRY/bin/gomobile build -target=ios $GITDIR/volley
ios-deploy --bundle volley.app
```

Just click to run, there's no further setup.

It the app finds the mounttable, it will join the game.
