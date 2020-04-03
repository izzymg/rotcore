/*
Package peer pertains to WebRTC Peer Connection management.
*/
package peer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"bitbucket.org/izzymg/rtmessages"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v2"
	"github.com/twitchtv/twirp"
)

// ErrNoSuchPeer occurs when a peer is looked up by an ID and not found.
var ErrNoSuchPeer = errors.New("No such peer")

var peerConfiguration = webrtc.Configuration{}

/*
New sets up and returns a RTC streamer.

The streamer acts as an ICE-Lite SFU, taking data from the given streams
and writing them to any peers it receives from remote procedure calls.

The public IPs are used to tell peers where the server is on the network,
or public internet, to avoid needing to do a STUN lookup. As such,
this server should be behind a DNAT public IP address.
*/
func New(publicIPs []string) (*Streamer, error) {

	var settings webrtc.SettingEngine
	settings.SetLite(true)
	settings.SetNAT1To1IPs(publicIPs, webrtc.ICECandidateTypeHost)
	settings.SetEphemeralUDPPortRange(11000, 13000)

	var me webrtc.MediaEngine
	me.RegisterCodec(h264Codec)

	server := &Streamer{
		api:   webrtc.NewAPI(webrtc.WithSettingEngine(settings), webrtc.WithMediaEngine(me)),
		peers: make(map[string]*webrtc.PeerConnection),
	}

	return server, nil
}

/*
Streamer manages and streams data to WebRTC peers using RPCs to signal with
the central service, storeing active peer connections.
*/
type Streamer struct {
	api    *webrtc.API
	peers  map[string]*webrtc.PeerConnection
	mut    sync.RWMutex
	stream mediadevices.MediaStream
}

// Handshake handles incoming offers from peers, returning back an answer.
func (s *Streamer) Handshake(ctx context.Context, incomingOffer *rtmessages.Offer) (*rtmessages.Answer, error) {
	log.Printf("Offer received from %q\n", incomingOffer.GetPeerId())

	peerID := incomingOffer.GetPeerId()
	if len(peerID) < 1 {
		return nil, twirp.RequiredArgumentError("Peer ID is required")
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  incomingOffer.GetSdp(),
	}

	// Create a connection to the peer.
	peer, err := s.api.NewPeerConnection(peerConfiguration)
	if err != nil {
		return nil, twirp.InternalErrorWith(fmt.Errorf("failed to create peer connection: %w", err))
	}

	// Add tracks to peer
	if s.stream == nil {
		stream, err := getStream(peer)
		if err != nil {
			return nil, twirp.InternalErrorWith(fmt.Errorf("failed to get stream: %w", err))
		}
		s.stream = stream
	}
	for _, tracker := range s.stream.GetTracks() {
		t := tracker.Track()
		tracker.OnEnded(func(err error) {
			log.Printf("Track (ID: %s, Label: %s) ended with error: %v\n",
				t.ID(), t.Label(), err)
		})
		_, err = peer.AddTransceiverFromTrack(t, webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		})
		if err != nil {
			return nil, twirp.InternalErrorWith(fmt.Errorf("failed to add track: %w", err))
		}
	}

	s.addPeer(peerID, peer)

	// Hook to remove peers when the connection state fails.
	peer.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed {
			s.removePeer(peerID)
		}
	})

	// Set description about the remote peer, then answer the peer
	err = peer.SetRemoteDescription(offer)
	if err != nil {
		s.removePeer(peerID)
		return nil, twirp.InternalErrorWith(fmt.Errorf("failed to set remote desc: %w", err))
	}

	answer, err := peer.CreateAnswer(nil)
	if err != nil {
		s.removePeer(peerID)
		return nil, twirp.InternalErrorWith(err)
	}

	err = peer.SetLocalDescription(answer)
	if err != nil {
		s.removePeer(peerID)
		return nil, twirp.InternalErrorWith(err)
	}

	return &rtmessages.Answer{
		PeerId: peerID,
		Sdp:    answer.SDP,
	}, nil
}

// NewCandidate pushes ICE candidates onto WebRTC peers.
func (s *Streamer) NewCandidate(ctx context.Context, incomingCandidate *rtmessages.Candidate) (*rtmessages.OK, error) {
	log.Printf("Candidate received from %q`\n", incomingCandidate.GetPeerId())

	peer := s.getPeer(incomingCandidate.GetPeerId())
	if peer == nil {
		return nil, twirp.NotFoundError("No such peer by that ID")
	}

	mid := incomingCandidate.GetSDPMid()
	index := uint16(incomingCandidate.GetSDPMLineIndex())

	peer.AddICECandidate(webrtc.ICECandidateInit{
		SDPMid:           &mid,
		SDPMLineIndex:    &index,
		Candidate:        incomingCandidate.GetCandidate(),
		UsernameFragment: incomingCandidate.GetUsernameFragment(),
	})

	return &rtmessages.OK{
		Ok: 1,
	}, nil
}

// Add a peer connection to the map, thread safe.
func (s *Streamer) addPeer(id string, peer *webrtc.PeerConnection) {
	s.mut.Lock()
	s.peers[id] = peer
	s.mut.Unlock()
	log.Printf("Added peer %q\n", id)
}

// Fetch a peer connection from the map, thread safe.
func (s *Streamer) getPeer(id string) *webrtc.PeerConnection {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.peers[id]
}

// Remove a peer connection from the map and close the underlying connection, thread safe.
func (s *Streamer) removePeer(id string) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	if peer, ok := s.peers[id]; ok {
		log.Printf("Removed peer %q\n", id)
		err := peer.Close()
		delete(s.peers, id)
		return err
	}
	return ErrNoSuchPeer
}
