/*
Package socket defines a websocket implementation of signalling.
*/
package socket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/izzymg/sockparty"
	"golang.org/x/time/rate"
)

// New spawns a new WebSocket signaling server.
func New(ctx context.Context, addr string) *Socketer {
	inc := make(chan sockparty.Incoming)
	join := make(chan uuid.UUID)

	party := sockparty.New("Socketer", inc, join, nil, &sockparty.Options{
		AllowCrossOrigin: true,
		PingFrequency:    0,
		RateLimiter:      rate.NewLimiter(rate.Every(time.Millisecond*80), 5),
	})

	server := &http.Server{
		Addr:    addr,
		Handler: party,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("Socketer shutting down")
			party.End("Context canceled")
			server.Shutdown(context.Background())
		}
	}()

	socketer := &Socketer{
		incoming:         inc,
		join:             join,
		party:            party,
		joinHandler:      func(id string) {},
		candidateHandler: func(i string, c []byte) {},
		answerHandler:    func(i string, a []byte) {},
		offerHandler:     func(i string, o []byte) {},
		restartHandler:   func() {},
	}

	go socketer.run(ctx)
	return socketer
}

// Socketer is a websocket implementation of signalling.
type Socketer struct {
	candidateHandler func(id string, cnd []byte)
	answerHandler    func(id string, ans []byte)
	offerHandler     func(id string, off []byte)
	restartHandler   func()
	joinHandler      func(id string)
	party            *sockparty.Party
	join             chan uuid.UUID
	incoming         chan sockparty.Incoming
}

func (s *Socketer) onCandidate(msg sockparty.Incoming) {
	s.candidateHandler(msg.UserID.String(), msg.Payload)
}

func (s *Socketer) onAnswer(msg sockparty.Incoming) {
	s.answerHandler(msg.UserID.String(), msg.Payload)
}

func (s *Socketer) onOffer(msg sockparty.Incoming) {
	s.offerHandler(msg.UserID.String(), msg.Payload)
}

// Starts the socketer's listening process.
func (s *Socketer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case userID := <-s.join:
			s.joinHandler(userID.String())
		case msg := <-s.incoming:
			switch msg.Event {
			case "restart":
				s.restartHandler()
			case "answer":
				s.onAnswer(msg)
			case "offer":
				s.onOffer(msg)
			case "candidate":
				s.onCandidate(msg)
			default:
				fmt.Printf("Invalid event %q\n", msg.Event)
			}
		}
	}
}

// RegisterAnswerHandler registers a callback when an answer SDP is received.
func (s *Socketer) RegisterAnswerHandler(callback func(id string, answerJSON []byte)) {
	s.answerHandler = callback
}

// RegisterOfferHandler registers a callback when an offer SDP is received.
func (s *Socketer) RegisterOfferHandler(callback func(id string, offerJSON []byte)) {
	s.offerHandler = callback
}

// RegisterICECandidateHandler registers a callback when a client has ICE candidate.
func (s *Socketer) RegisterICECandidateHandler(callback func(id string, candidateJSON []byte)) {
	s.candidateHandler = callback
}

// SendICECandidate sends ICE candidate to a peer.
func (s *Socketer) SendICECandidate(id string, candidateJSON []byte) error {
	return s.party.Message(
		context.Background(),
		uuid.MustParse(id),
		&sockparty.Outgoing{
			Event:   "candidate",
			Payload: candidateJSON,
		})
}

// SendOffer sends the peer at ID an offer.
func (s *Socketer) SendOffer(id string, offer []byte) error {
	return s.party.Message(context.Background(), uuid.MustParse(id), &sockparty.Outgoing{
		Event:   "offer",
		Payload: offer,
	})
}

// SendAnswer sends the peer at ID an answer.
func (s *Socketer) SendAnswer(id string, offer []byte) error {
	return s.party.Message(context.Background(), uuid.MustParse(id), &sockparty.Outgoing{
		Event:   "answer",
		Payload: offer,
	})
}

// RegisterRestartHandler registers a callback on restart request received.
func (s *Socketer) RegisterRestartHandler(callback func()) {
	s.restartHandler = callback
}

// RegisterJoinHandler registers a callback on a new peer joining.
func (s *Socketer) RegisterJoinHandler(callback func(id string)) {
	s.joinHandler = callback
}
