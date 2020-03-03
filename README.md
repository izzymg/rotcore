# RotCore WebRTC & X11 streamer

This repository is made up of three components used by rooms for RotCore.

At the top level is **webrtc_send**, a **Go** module which acts as an SFU and WebRTC peer connector,
signaling SDPs through the given signaling program, receiving video from **xsend** over UDP.

**xsend** is written in **C/GStreamer**, used to stream the desktop video & audio from X11/Pulse using GStreamer, to UDP ports on the system.

**kbm** is written in **Rust**, used to control the keyboard and mouse of the desktop, receiving information via a TCP stream.

## Build & Bootstrap

The easiest way to run these components is by using the **NodeJS** bootstrap script.

Copy the secret into `conf/secret`, which is ignored by git.

Run `make all` to trigger a build of the project, then run `SIGNAL_ADDRESS=localhost:80 node boostrap.js` to spawn instances of:

* X11 & Chromium

* XSend

* KBM

* WebRTCSend

## WebRTCSend dependencies

To tell git to use your SSH keys to authenticate with Github:

`git config url."git@github.com:".insteadOf "https://github.com/"`

You'll also want in your `.bashrc` or similar:

`export GOPRIVATE="$GOPRIVATE,github.com/izzymg"`

Then:

`go mod download`

`go 1.13+`

#### Environment

`SIGNAL_ADDRESS` Address, without protocol, for signaling

`AUDIO_STREAM_ADDRESS` UDP address of audio stream data

`VIDEO_STREAM_ADDRESS` UDP address of video stream data

## Streamer dependencies

`gcc`

`pkg-config`

`gstreamer 1.10+`

`gstreamer-plugins`: base,good,bad,ugly

#### Environment

...

## KBM dependencies

Rust: See `cargo.toml`

## Args

1. Address of TCP server