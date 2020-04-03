package peer

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/frame"

	// Required for screen capture
	_ "github.com/pion/mediadevices/pkg/driver/screen"
	"github.com/pion/webrtc/v2"
)

// Returns x264 codec parameters.
func getx264Parameters() *x264.Params {
	mParams, err := x264.NewParams()
	if err != nil {
		panic(fmt.Errorf("failed to get codec parameters: %w", err))
	}
	mParams.Preset = x264.PresetVeryfast
	mParams.BitRate = 5000 * 1000
	return &mParams
}

// Defines video track capacities
func videoConstraint(c *mediadevices.MediaTrackConstraints) {
	c.Enabled = true
	c.FrameFormat = frame.FormatI420
	c.VideoEncoderBuilders = []codec.VideoEncoderBuilder{getx264Parameters()}
	c.FrameRate = 30
	c.Height = 1280
	c.Width = 720
}

// Returns the main media stream.
func getStream(peer *webrtc.PeerConnection) (mediadevices.MediaStream, error) {
	md := mediadevices.NewMediaDevices(peer)
	media, err := md.GetDisplayMedia(mediadevices.MediaStreamConstraints{
		Video: videoConstraint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get display media: %w", err)
	}
	return media, nil
}

var opusCodec = webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
var h264Codec = webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000)
