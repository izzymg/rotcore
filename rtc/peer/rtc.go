/*
Package peer pertains to WebRTC Peer Connection management.
*/
package peer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/izzymg/rotcommon/rtcservice"
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
func New(streams []Stream, publicIPs []string) (*Streamer, error) {

	settings := webrtc.SettingEngine{}
	settings.SetLite(true)
	settings.SetNAT1To1IPs(publicIPs, webrtc.ICECandidateTypeHost)

	medias := webrtc.MediaEngine{}
	medias.RegisterCodec(h264Codec)
	medias.RegisterCodec(opusCodec)

	// Initialize all tracks -> streams
	trackStreams, err := makeTrackStreams(streams)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize track streams: %w", err)
	}

	server := &Streamer{
		api:          webrtc.NewAPI(webrtc.WithMediaEngine(medias), webrtc.WithSettingEngine(settings)),
		peers:        make(map[string]*webrtc.PeerConnection),
		trackStreams: trackStreams,
	}

	return server, nil
}

/*
Streamer manages and streams data to WebRTC peers using RPCs to signal with
the central service, storeing active peer connections.
*/
type Streamer struct {
	api          *webrtc.API
	peers        map[string]*webrtc.PeerConnection
	mut          sync.RWMutex
	trackStreams map[*webrtc.Track]Stream
}

/*
StartStreaming begins streaming data from configured UDP streams to the RTC tracks,
spawning len(streams) goroutines.
*/
func (s *Streamer) StartStreaming(ctx context.Context) error {
	for track, stream := range s.trackStreams {
		raddr, err := net.ResolveUDPAddr("udp", stream.UDPAddress)
		if err != nil {
			return err
		}
		conn, err := net.ListenUDP("udp", raddr)
		if err != nil {
			return err
		}

		var writeStream func()
		switch stream.Type {
		case H264Stream:
			writeStream = writeH264(track, conn)
		case OpusStream:
			writeStream = writeOpus(track, conn)
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Stream stopping: %v\n", ctx.Err())
					conn.Close()
				default:
					writeStream()
				}
			}
		}()
	}
	return nil
}

// Handshake handles incoming offers from peers, returning back an answer.
func (s *Streamer) Handshake(ctx context.Context, incomingOffer *rtcservice.Offer) (*rtcservice.Answer, error) {
	fmt.Printf("Offer received from %q\n", incomingOffer.GetPeerId())

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
		return nil, twirp.InternalErrorWith(err)
	}

	s.addPeer(peerID, peer)

	// Hook to remove peers when the connection state fails.
	peer.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed {
			s.removePeer(peerID)
		}
	})

	// Add the stream tracks to send on our connection
	for track := range s.trackStreams {
		_, err := peer.AddTrack(track)
		if err != nil {
			s.removePeer(peerID)
			return nil, twirp.InternalErrorWith(err)
		}
	}

	// Set description about the remote peer, then answer the peer

	err = peer.SetRemoteDescription(offer)
	if err != nil {
		s.removePeer(peerID)
		return nil, twirp.InternalErrorWith(err)
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

	return &rtcservice.Answer{
		PeerId: peerID,
		Sdp:    answer.SDP,
	}, nil
}

// NewCandidate pushes ICE candidates onto WebRTC peers.
func (s *Streamer) NewCandidate(ctx context.Context, incomingCandidate *rtcservice.Candidate) (*rtcservice.OK, error) {
	fmt.Printf("Candidate received from %q`\n", incomingCandidate.GetPeerId())

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

	return &rtcservice.OK{
		Ok: 1,
	}, nil
}

// Add a peer connection to the map, thread safe.
func (s *Streamer) addPeer(id string, peer *webrtc.PeerConnection) {
	s.mut.Lock()
	s.peers[id] = peer
	s.mut.Unlock()
	fmt.Printf("Added peer %q\n", id)
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
		fmt.Printf("Removed peer %q\n", id)
		err := peer.Close()
		delete(s.peers, id)
		return err
	}
	return ErrNoSuchPeer
}
