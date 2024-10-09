package video

import (
	"fmt"
	"motionspeed/internal/frame"

	"gocv.io/x/gocv"
)

type Stream struct {
	Video *gocv.VideoCapture
}

func NewStream(videoPath string) (*Stream, error) {
	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open video file: %v", err)
	}
	return &Stream{Video: video}, nil
}

func (s *Stream) Close() {
	s.Video.Close()
}

func (s *Stream) Fps() float64 {
	return s.Video.Get(gocv.VideoCaptureFPS)
}

func (s *Stream) TimeAtFrame(frame *frame.Frame) float64 {
	return float64(frame.FrameIndex()) / s.Fps()
}
