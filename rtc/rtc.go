/*
Package rtc pertains to WebRTC Peer Connection management.
*/
package rtc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/izzymg/rotcommon"

	"github.com/pion/webrtc/v2"
)

// ErrNoSuchPeer occurs when a peer is looked up by an ID and not found.
var ErrNoSuchPeer = errors.New("No such peer")

var peerConfiguration = webrtc.Configuration{}

// NewServer returns a server using the given signaler.
func NewServer(signaler rotcommon.RTCBridge, streams []Stream) (*Server, error) {
	// Allow ICE trickling & ICE LITE
	settings := webrtc.SettingEngine{}
	settings.SetLite(true)
	medias := webrtc.MediaEngine{}
	// Create and register supported codecs.
	medias.RegisterCodec(h264Codec)
	medias.RegisterCodec(opusCodec)

	// Initialize all tracks -> streams
	trackStreams, err := makeTrackStreams(streams)
	if err != nil {
		return nil, err
	}

	// Server setup with signaler
	server := &Server{
		api:          webrtc.NewAPI(webrtc.WithMediaEngine(medias), webrtc.WithSettingEngine(settings)),
		signaler:     signaler,
		peers:        make(map[string]*webrtc.PeerConnection),
		trackStreams: trackStreams,
	}

	signaler.RegisterOfferHandler(server.onOffer)
	signaler.RegisterICECandidateHandler(server.onICECandidate)

	return server, nil
}

/*
Server manages and streams data to WebRTC peers using a given signaler.
The signaler passes offers/candidates and so on to the RTC server,
which then uses the information to establish direct peer connections.
It stores connected peers along with their signaler ID.
*/
type Server struct {
	api          *webrtc.API
	signaler     rotcommon.RTCBridge
	peers        map[string]*webrtc.PeerConnection
	mut          sync.RWMutex
	trackStreams map[*webrtc.Track]Stream
}

/*
StartStreaming begins streaming data from configured UDP streams to the RTC tracks.
Spawns len(streams) goroutines.
*/
func (s *Server) StartStreaming(ctx context.Context) error {
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

// Add a peer connection to the list
func (s *Server) addPeer(id string, peer *webrtc.PeerConnection) {
	s.mut.Lock()
	s.peers[id] = peer
	s.mut.Unlock()
	fmt.Printf("Added peer %q\n", id)
}

// Remove a peer connection from the list and close the underlying connection
func (s *Server) removePeer(id string) error {
	s.mut.Lock()
	if peer, ok := s.peers[id]; ok {
		fmt.Printf("Removed peer %q\n", id)
		err := peer.Close()
		delete(s.peers, id)
		s.mut.Unlock()
		return err
	}
	s.mut.Unlock()
	return ErrNoSuchPeer
}

// TODO: drop peer from list if any failures during offer-answer setup

// Handle received offers by the singaler, returning back an answer if applicable.
func (s *Server) onOffer(id string, offerJSON []byte) {

	offer := webrtc.SessionDescription{}
	err := json.Unmarshal(offerJSON, &offer)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Offer received from %q\n", id)

	// Create a connection to the peer.
	// Peer is added so incoming ICE candidates know where to go.
	peer, err := s.api.NewPeerConnection(peerConfiguration)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.addPeer(id, peer)

	// Register handlers
	peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Send new ICE candidates to the client
			fmt.Println("Sending ICE candidate")
			candidateJSON, err := json.Marshal(candidate.ToJSON())
			if err != nil {
				fmt.Printf("Error marshaling candidate: %+v", err)
			}
			s.signaler.SendICECandidate(id, candidateJSON)
		} else {
			fmt.Println("Skipping nil candidate")
		}
	})

	peer.OnSignalingStateChange(func(state webrtc.SignalingState) {
		fmt.Printf("Signal State change: %+v\n", state)
	})
	peer.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE State change: %+v\n", state)
		if state == webrtc.ICEConnectionStateFailed {
			peer.Close()
			s.removePeer(id)
		}
	})

	// Add the stream tracks to send on our connection
	for track := range s.trackStreams {
		_, err := peer.AddTrack(track)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Set description about the remote peer, then answer the peer
	err = peer.SetRemoteDescription(offer)
	if err != nil {
		fmt.Println(err)
		return
	}

	answer, err := peer.CreateAnswer(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = peer.SetLocalDescription(answer)
	if err != nil {
		fmt.Println(err)
		return
	}

	answerJSON, err := json.Marshal(answer)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.signaler.SendAnswer(id, answerJSON)
}

// Handler received an ICE candidate by the singaler, add to peer connection
func (s *Server) onICECandidate(id string, candidateJSON []byte) {
	s.mut.Lock()
	defer s.mut.Unlock()
	if peer, ok := s.peers[id]; ok {
		candidateInit := webrtc.ICECandidateInit{}
		err := json.Unmarshal(candidateJSON, &candidateInit)
		if err != nil {
			fmt.Println(err)
			return
		}
		peer.AddICECandidate(candidateInit)
		fmt.Printf("Added ICE candidate from %q\n", id)
	} else {
		fmt.Printf("Unknown candidate peer destination %q\n", id)
	}
}

func (s *Server) onRestart() {
	fmt.Println("RESTART UNIMPLEMENTED")
}
