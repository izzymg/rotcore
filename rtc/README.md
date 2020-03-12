### RTC

RTC is the Golang application which responds to WebRTC peer calls, and streams the RTC desktop stream to them.

It runs an HTTP server using `Twirp` to respond to remote procedure calls from authorized applications.

It takes a secret used to authenticate requests, and binds to arbitrary UDP ports to stream.

It takes stream data from `../streamer` on UDP ports specified by `AUDIO_STREAM_ADDRESS` and `VIDEO_STREAM_ADDRESS`.
These should be `Opus` & `h264 constrained-baseline byte-stream` encoded streams, respectively.

### Public IP addresses behind NAT

As RTC performs p2p connections, it needs to advertise its Public IP address to the peers. 

`rtc --ip=xxx.xxx.xx.xx --ip=yyy.yyy.yy.yy --secret=secret.txt`