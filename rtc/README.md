### RTC

RTC is the Golang application which responds to WebRTC peer calls, and streams the RTC desktop stream to them.

It runs an HTTP server using `Twirp` to respond to remote procedure calls from authorized applications.

It takes a secret used to authenticate requests, and binds to arbitrary UDP ports to stream.


### Public IP addresses behind NAT

As RTC performs p2p connections, it needs to advertise its Public IP address to the peers. 

`rtc --ip=xxx.xxx.xx.xx --ip=yyy.yyy.yy.yy --secret=secret.txt`

### Env

`ROTCORE_ADDRESS=0.0.0.0:3001`

### Deps

`apt-get install libx11-dev libxext-dev`