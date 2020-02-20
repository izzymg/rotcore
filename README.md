# RotCore WebRTC & X11 streamer

This repository is made up of three components used by rooms for RotCore.

At the top level is **webrtc_send**, a Go module which acts as an SFU and WebRTC peer connector,
signaling SDPs through the given signaling program, receiving video from **xsend** over UDP.

**xsend** is used to stream the desktop video & audio from X11/Pulse using GStreamer, to UDP ports on the system.

**xinteract** is used to control the keyboard and mouse of the desktop, receiving information over a message queue via TCP.

## Build & Bootstrap

The easiest way to run these components is by using the **NodeJS** bootstrap script.

Copy the secret into `conf/secret`, which is ignored by git.

Run `make all` to trigger a build of the project, then run `OTHER_BRIDGE_ADDRESS=http://example.com/bridge node boostrap.js` to spawn instances of:

* X11 & Chromium

* XSend

* XInteract

* WebRTCSend

## WebRTC_Send dependencies

To tell git to use your SSH keys to authenticate with Github:

`git config url."git@github.com:".insteadOf "https://github.com/"`

You'll also want in your `.bashrc` or similar:

`export GOPRIVATE="$GOPRIVATE,github.com/izzymg"`

Then:

`go mod download`

`go 1.13+`

#### Environment

`SIGNAL_ADDRESS` Address, without protocol, for signaling

`SIGNAL_WITH_SOCKET` Use direct WebSockets for signaling

`AUDIO_STREAM_ADDRESS` UDP address of audio stream data

`VIDEO_STREAM_ADDRESS` UDP address of video stream data

`OTHER_BRIDGE_ADDRESS` Address of the HTTP Bridge if used for signaling

`SECRET_PATH` Path of the secret for verification of signaling requests


## XSend dependencies

`gcc`

`pkg-config`

`gstreamer 1.10+`

`gstreamer-plugins`: base,good,bad,ugly

#### Environment

...

## XInteract dependencies

`libczmq libzmq 4.3+`

`xdo.h` see: xdotool

#### Environment
`XI_USERNAME` Username used to authenticate with publisher

`XI_PASSWORD` Password used to authenticate with publisher

`XI_ADDRESS="tcp://127.0.0.1"` > `xinteract tcp://128.0.0.1`

