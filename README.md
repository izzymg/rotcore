# Rotcore

This repository acts as a monorepo for all the source code of all of the programs needed to run a Rotcore **room**.

### RTC

RTC is the Golang application which responds to WebRTC peer calls, and streams the RTC desktop stream to them.

It runs an HTTP server using `Twirp` to respond to remote procedure calls from authorized applications.

It takes a secret used to authenticate requests, and binds to arbitrary UDP ports to stream.

### KBM

KBM is a Rust program which allows interaction with the mouse & keyboard in x11. It connects to the x11 `$DISPLAY`,
and starts a TCP server which authenticates only one single TCP stream if it provides an acceptable challenge.

That server responds to special commands to click, scroll, type, etc.

As such, it will want the same secret as RTC.

### Streamer

Streamer a tiny C program which takes a gstreamer pipeline string as an argument and runs it.

In the future, it is planned to automatically restart, log events such as backlogged queues, etc.


### Provision

Provision contains some data used to actually build a room virtual machine from scratch.

### Deploy

A series of programs is also needed to make Rotcore useful, including an X11 display, PulseAudio, Firefox,
and their related configuration files for usability and security.

The `deploy` folder contains a series of scripts which run all of the aforementioned applications with
specific flags and settings, culminating in `deploy.js` which spawns RTC, KBM, Streamer, X11, and Firefox, 
all from one easy to run NodeJS script. Configuration is done by copying `config.ex.js` -> `config.js`.

### Building

Before running `deploy.js`, a build needs to be done of the 3 executables. There's a handy Dockerfile and associated
`docker_build.sh` that can do this. It takes `GIT_TOKEN` as an environment variable, builds RTC, KBM, Streamer, and copies
the resulting executables into `./deploy/bin` to be run by `deploy.js`

#### TL;DR

```shell
pacman -S nodejs firefox xorg-server pulseaudio # v8+
GIT_TOKEN=aaabbbcccdddeeefff docker_build.sh
cd deploy
node deploy.js
```