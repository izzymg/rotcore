package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rtc/auth"
	"rtc/peer"
	"strings"
	"time"

	"github.com/izzymg/rotcommon/rtcservice"
)

/*
RotCore WebRTC entrypoint - IzzyMG
Acts as an RPC server to receive SDPs and send stream data to peers.
*/

/*
Joins multiple flags into a slice.
e.g. --user "mary" --user "alice" --user "bob".
*/
type multiStringFlag []string

func (af *multiStringFlag) String() string {
	return strings.Join(*af, ", ")
}

func (af *multiStringFlag) Set(value string) error {
	*af = append(*af, value)
	return nil
}

func main() {

	secretPath := flag.String("secret", "./secret.txt", "Path to a file containing a secret used to verify requests")

	var publicIps multiStringFlag
	flag.Var(&publicIps, "ip", "DNAT public IPs")

	flag.Parse()

	fmt.Printf("WebRTC ICE-Lite IPs: %s\n", publicIps.String())

	// Setup main context and configuration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr, ok := os.LookupEnv("SIGNAL_ADDRESS")
	if !ok {
		addr = ":80"
	}

	videoStreamAddr, ok := os.LookupEnv("VIDEO_STREAM_ADDRESS")
	if !ok {
		videoStreamAddr = "127.0.0.1:9577"
	}

	audioStreamAddr, ok := os.LookupEnv("AUDIO_STREAM_ADDRESS")
	if !ok {
		audioStreamAddr = "127.0.0.1:9578"
	}

	// Setup RTC server, which implements the RPC protos.
	streamer, err := peer.New([]peer.Stream{
		peer.Stream{
			Type:       peer.OpusStream,
			UDPAddress: audioStreamAddr,
		},
		peer.Stream{
			Type:       peer.H264Stream,
			UDPAddress: videoStreamAddr,
		},
	}, publicIps)

	if err != nil {
		cancel()
		panic(err)
	}

	// Begin streaming from UDP streams.

	err = streamer.StartStreaming(ctx)
	// Ignore context cancel errors before panics
	if err != nil && ctx.Err() == nil {
		cancel()
		panic(err)
	}

	// Initialize RPC server, using HTTP auth middleware as handler.

	fmt.Printf("Starting RTC's RPC server on %s\n", addr)
	server := rtcservice.NewRTCServer(streamer, nil)

	auth, err := auth.Middleware(server, *secretPath)
	if err != nil {
		panic(err)
	}

	httpServer := http.Server{
		Handler: auth,
		Addr:    addr,
	}

	go httpServer.ListenAndServe()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		fmt.Println("Got interrupt")
		httpServer.Close()
		cancel()
		<-time.After(300 * time.Millisecond)
		return
	}
}
