package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"webrtc_send/auth"

	"github.com/izzymg/rotcommon"
	"github.com/izzymg/rotcommon/rtcservice"
	"github.com/matryer/is"
	"github.com/twitchtv/twirp"
)

/*
Client side tests for Rotcore's WebRTC component.
*/

type stubService struct {
	OnHandshake func()
}

func (s *stubService) Handshake(context.Context, *rtcservice.Offer) (*rtcservice.Answer, error) {
	s.OnHandshake()
	return &rtcservice.Answer{
		PeerId: "AAAA",
	}, nil
}
func (s *stubService) NewCandidate(context.Context, *rtcservice.Candidate) (*rtcservice.OK, error) {
	return &rtcservice.OK{}, nil
}

/*
Runs a series of tests to check that only authenticated users can
make requests to the RPC server.
TODO: refactor this out into short, mappable tests instead of one big function.
*/
func TestAuthorization(t *testing.T) {

	is := is.New(t)
	addr := "localhost:9954"
	secret := "iamfallingiamfading"

	/* Setup server */
	service := &stubService{}
	server := rtcservice.NewRTCServer(service, nil)

	httpServer := http.Server{
		Handler: auth.Middleware(server, secret),
		Addr:    addr,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			httpServer.Shutdown(context.Background())
		}
	}(ctx)

	go httpServer.ListenAndServe()

	/* Setup client */
	client := rtcservice.NewRTCProtobufClient("http://"+addr, &http.Client{})

	// Setup headers without the token
	header := make(http.Header)
	header.Set(auth.AuthHeader, "cute-days")
	ctx, err := twirp.WithHTTPRequestHeaders(ctx, header)
	if err != nil {
		is.Fail()
	}

	// Expect an error complaining about the missing token
	_, err = client.Handshake(ctx, &rtcservice.Offer{PeerId: "garbage"})
	if err == nil {
		is.Fail()
	}

	twirpErr, ok := err.(twirp.Error)
	if !ok {
		is.Fail()
	}

	body := twirpErr.Meta("body")

	is.Equal(twirpErr.Code(), twirp.PermissionDenied)
	is.Equal(body, auth.MissingTokenMessage)

	// Setup headers without the auth
	header = make(http.Header)
	header.Set(auth.TokenHeader, "cute-days")
	ctx, err = twirp.WithHTTPRequestHeaders(ctx, header)
	if err != nil {
		is.Fail()
	}

	// Expect an error complaining about the missing auth
	_, err = client.Handshake(ctx, &rtcservice.Offer{PeerId: "garbage"})
	if err == nil {
		is.Fail()
	}

	twirpErr, ok = err.(twirp.Error)
	if !ok {
		is.Fail()
	}

	body = twirpErr.Meta("body")

	is.Equal(twirpErr.Code(), twirp.PermissionDenied)
	is.Equal(body, auth.MissingAuthMessage)

	// Setup headers with an incorrect hash
	header = make(http.Header)
	digest, err := rotcommon.HashData("cute-days", "some-garbage")
	if err != nil {
		is.Fail()
	}
	header.Set(auth.TokenHeader, "cute-days")
	header.Set(auth.AuthHeader, digest)

	ctx, err = twirp.WithHTTPRequestHeaders(ctx, header)
	if err != nil {
		is.Fail()
	}

	// Expect a permission denied error
	_, err = client.Handshake(ctx, &rtcservice.Offer{PeerId: "garbage"})
	if err == nil {
		is.Fail()
	}

	twirpErr, ok = err.(twirp.Error)
	if !ok {
		is.Fail()
	}
	is.Equal(twirpErr.Code(), twirp.PermissionDenied)

	// Finally try with a correct hash

	var wg sync.WaitGroup
	wg.Add(1)
	service.OnHandshake = func() {
		wg.Done()
	}

	header = make(http.Header)
	digest, err = rotcommon.HashData("cute-days", secret)
	if err != nil {
		is.Fail()
	}
	header.Set(auth.TokenHeader, "cute-days")
	header.Set(auth.AuthHeader, digest)

	ctx, err = twirp.WithHTTPRequestHeaders(ctx, header)
	if err != nil {
		is.Fail()
	}

	// Expect a permission denied error
	_, err = client.Handshake(ctx, &rtcservice.Offer{PeerId: "garbage"})
	if err != nil {
		is.Fail()
	}

	wg.Wait()

	cancel()
}
