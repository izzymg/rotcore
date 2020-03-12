package peer

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

/*
	Data -> track writers for WebRTC. izzymg 2020.
*/
var opusCodec = webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
var h264Codec = webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000)

// Stream types
const (
	H264Stream = uint8(iota)
	OpusStream
)

// Stream represents a UDP stream of data pulled into peers.
type Stream struct {
	UDPAddress string
	Type       uint8
}

// Maps all stream types & addresses given to new WebRTC tracks.
func makeTrackStreams(streams []Stream) (map[*webrtc.Track]Stream, error) {
	trackStreams := make(map[*webrtc.Track]Stream)
	for _, stream := range streams {
		switch stream.Type {
		case H264Stream:
			videoTrack, err := webrtc.NewTrack(
				webrtc.DefaultPayloadTypeH264,
				rand.Uint32(),
				"video",
				"room-video",
				h264Codec,
			)
			if err != nil {
				return nil, err
			}
			trackStreams[videoTrack] = stream
		case OpusStream:
			audioTrack, err := webrtc.NewTrack(
				webrtc.DefaultPayloadTypeOpus,
				rand.Uint32(),
				"audio",
				"room-audio",
				opusCodec,
			)
			if err != nil {
				return nil, err
			}
			trackStreams[audioTrack] = stream
		}
	}
	return trackStreams, nil
}

// Generates an returns a function which writes H264 data from reader into track.
func writeH264(track *webrtc.Track, conn io.Reader) func() {
	buf := make([]byte, 1024*512)
	rate := int64(track.Codec().ClockRate / 1000)
	timeout := time.Second * 5

	return func() {
		/* Read in n bytes from source, calculate samples to push onto track
		and write them in in a loop. */
		start := time.Now()
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Stream read error, trying again in %f s: %v\n",
				timeout.Seconds(),
				err,
			)
			<-time.After(timeout)
		}

		// TODO: use upcoming track.Packetizer() method to write raw RTP packets
		track.WriteSample(media.Sample{
			Data:    buf[:n],
			Samples: uint32(rate * time.Since(start).Milliseconds()),
		})
	}
}

// Generates an returns a function which writes OPUS data from reader into track.
func writeOpus(track *webrtc.Track, conn io.Reader) func() {
	buf := make([]byte, 1024)
	rate := uint32(track.Codec().ClockRate / 1000)
	frameSize := uint32(10)
	timeout := time.Second * 5

	return func() {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Stream read error, trying again in %f s: %v\n",
				timeout.Seconds(),
				err,
			)
			<-time.After(timeout)
		}
		// TODO: use upcoming track.Packetizer() method to write raw RTP packets
		track.WriteSample(media.Sample{
			Data:    buf[:n],
			Samples: rate * frameSize,
		})
	}
}
