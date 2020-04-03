package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rtc/auth"
	"rtc/peer"
	"strings"
	"time"

	"bitbucket.org/izzymg/rtmessages"
)

/*
RotCore WebRTC entrypoint - IzzyMG
Acts as an RPC server to receive SDPs and send stream data to peers. */

/*
Joins multiple flags into a slice.
e.g. --user "mary" --user "alice" --user "bob". */
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
	insecure := flag.Bool("insecure", false, "Don't verify incoming requests.")

	var publicIps multiStringFlag
	flag.Var(&publicIps, "ip", "DNAT public IPs")

	flag.Parse()

	fmt.Printf("WebRTC ICE-Lite IPs: %s\n", publicIps.String())

	addr, ok := os.LookupEnv("ROTCORE_ADDRESS")
	if !ok {
		addr = ":80"
	}
	// Setup RTC server, which implements the RPC protos.
	streamer, err := peer.New(publicIps)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting RTC's RPC server on %s\n", addr)
	server := rtmessages.NewRTCServer(streamer, nil)

	// Serve auth middleware if server is secured.
	auth, err := auth.Middleware(server, *secretPath)
	if err != nil {
		panic(err)
	}
	httpServer := http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if *insecure == false {
				auth.ServeHTTP(rw, req)
				return
			}
			server.ServeHTTP(rw, req)
		}),
		Addr: addr,
	}

	go httpServer.ListenAndServe()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		fmt.Println("Got interrupt")
		httpServer.Close()
		<-time.After(300 * time.Millisecond)
		return
	}
}
