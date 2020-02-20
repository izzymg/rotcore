package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"webrtc_send/rtc"

	"github.com/izzymg/rotcommon/rtcservice"
)

// RotCore WebRTC entrypoint - IzzyMG, 2020

func main() {

	// Setup main context and configuration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr, ok := os.LookupEnv("SIGNAL_ADDRESS")
	if !ok {
		addr = ":80"
	}

	audioStreamAddr, ok := os.LookupEnv("AUDIO_STREAM_ADDRESS")
	if !ok {
		audioStreamAddr = "127.0.0.1:9577"
	}

	videoStreamAddr, ok := os.LookupEnv("VIDEO_STREAM_ADDRESS")
	if !ok {
		audioStreamAddr = "127.0.0.1:9578"
	}

	// Setup RTC server give it UDP streams.
	streamer, err := rtc.New([]rtc.Stream{
		rtc.Stream{
			Type:       rtc.OpusStream,
			UDPAddress: audioStreamAddr,
		},
		rtc.Stream{
			Type:       rtc.H264Stream,
			UDPAddress: videoStreamAddr,
		},
	})

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

	// Initialize RPC server

	fmt.Printf("Using %s\n", addr)
	server := rtcservice.NewRTCServer(streamer, nil)
	go http.ListenAndServe(addr, server)

	// Listen for interupts
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		fmt.Println("Got interrupt")
		cancel()
		<-time.After(300 * time.Millisecond)
		return
	}
}
