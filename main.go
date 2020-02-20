package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"webrtc_send/rtc"
	"webrtc_send/socket"

	"github.com/izzymg/rotcommon"
)

// RotCore WebRTC entrypoint - IzzyMG, 2020

func main() {

	// Setup main context and configuration
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr, ok := os.LookupEnv("SIGNAL_ADDRESS")
	if !ok {
		addr = "0.0.0.0:9952"
	}

	useSocket := os.Getenv("SIGNAL_WITH_SOCKET")

	audioStreamAddr, ok := os.LookupEnv("AUDIO_STREAM_ADDRESS")
	if !ok {
		audioStreamAddr = "127.0.0.1:5592"
	}

	videoStreamAddr, ok := os.LookupEnv("VIDEO_STREAM_ADDRESS")
	if !ok {
		audioStreamAddr = "127.0.0.1:5591"
	}

	var signaler rotcommon.RTCBridge

	if len(useSocket) > 0 && useSocket != "FALSE" && useSocket != "0" {
		// Use the socket server to stream
		signaler = socket.New(ctx, addr)
	} else {
		// Gateway just gives us an implementation and an HTTP router, no server.
		gw, err := rotcommon.New(
			os.Getenv("OTHER_BRIDGE_ADDRESS"),
			os.Getenv("SECRET_PATH"),
		)
		gw.Debug = true
		if err != nil {
			panic(err)
		}
		signaler = gw

		// Create server, use the gateway's router
		server := http.Server{
			Handler:           gw.Router,
			Addr:              addr,
			IdleTimeout:       time.Hour,
			ReadHeaderTimeout: time.Second * 20,
			ReadTimeout:       time.Second * 100,
			WriteTimeout:      time.Second * 100,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				server.Shutdown(ctx)
			}
		}()
	}
	fmt.Printf("Signaling server running on %q\n", addr)

	// Setup RTC server with a signaler and give it UDP streams.
	rtcServer, err := rtc.NewServer(signaler, []rtc.Stream{
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

	err = rtcServer.StartStreaming(ctx)
	// Ignore context cancel errors before panics
	if err != nil && ctx.Err() == nil {
		cancel()
		panic(err)
	}

	// Listen for interupts
	sigint := make(chan os.Signal, 1) // Must be buffered
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		fmt.Println("Got interrupt")
		cancel()
		<-time.After(300 * time.Millisecond)
		return
	}
}
