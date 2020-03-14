#/bin/bash

# Warning: This script expects the "stream" binary to exist in ../bin/stream !

# Runs streamer with an h.264 & Opus pipeline pulling in X11 & Pulse audio data.
# Pushes streams onto UDP sinks.

# The data is read from the RTC Go app to be streamed to clients over WebRTC.
# As such, this pipeline must produce browser-playable H264 + Opus data.
# The RTP packatizing is taken care of by the RTC application.

WIDTH=1280
HEIGHT=720
FRAMERATE=30/1

HOST=127.0.0.1

VIDEO_PORT=9577
AUDIO_PORT=9578

VIDEO_PRESET="fast"
VIDEO_BITRATE=750

echo "Streamer: $WIDTH x $HEIGHT at $FRAMERATE FPS"

pulseaudio --start

# H264 must used "constrained-baseline" and "byte-stream" for browser compatibility.
# 'zerolatency' is also important for the best quality stream.

# Opus is significantly less touchy, requiring only a clockrate and format to produce ok audio.

exec ../bin/streamer "ximagesrc use-damage=false ! video/x-raw,framerate=$FRAMERATE
! videoconvert ! videoscale ! video/x-raw,width=$WIDTH,height=$HEIGHT
! x264enc tune=zerolatency bitrate=$VIDEO_BITRATE speed-preset=$VIDEO_PRESET
! video/x-h264,stream-format=byte-stream,profile=constrained-baseline
! queue ! udpsink host=$HOST port=$VIDEO_PORT

pulsesrc ! audio/x-raw,format=S16LE,rate=48000,channels=2
! opusenc frame-size=10 audio-type=generic
! udpsink host=$HOST port=$AUDIO_PORT"
