# RotCore WebRTC & X11 streamer

## WebRTC dependencies

`gcc`

`pkg-config`

`glib-2.0`

`libxdo`

`go`

### Go Module/Git settings

To tell git to use your SSH keys to authenticate with Github:

`git config url."git@github.com:".insteadOf "https://github.com/"`

You'll also want in your `.bashrc` or similar:

`export GOPRIVATE="$GOPRIVATE,github.com/izzymg"`

Then:

`go mod download`

## Desktop streamer dependencies

`gcc`

`pkg-config`

`gstreamer 1.10+`

`gstreamer-plugins`: base,good,bad,ugly

## Building

`go mod download`

`CGO_ENABLED=1 go build -o webrtc`

`make`

## Environment

`SIGNAL_ADDRESS` Address, without protocol, for signaling

`SIGNAL_WITH_SOCKET` Use direct WebSockets for signaling

`AUDIO_STREAM_ADDRESS` UDP address of audio stream data

`VIDEO_STREAM_ADDRESS` UDP address of video stream data

`OTHER_GATEWAY_ADDRESS` Address of the HTTP Gateway if used for signaling

`GATEWAY_SECRET_PATH` Path of the secret for signaling
