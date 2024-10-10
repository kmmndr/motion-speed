package video

import (
	"fmt"
	"motionspeed/internal/frame"

	"gocv.io/x/gocv"
)

type Stream struct {
	Video *gocv.VideoCapture
}

func NewFileStream(videoPath string) (*Stream, error) {
	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open video file: %v", err)
	}
	return &Stream{Video: video}, nil
}

func NewDeviceStream(videoUrl string) (*Stream, error) {
	video, err := gocv.OpenVideoCapture(videoUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to open video url: %v", err)
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

func (s *Stream) Read(frameIndex int) *frame.Frame {
	currentFrame := gocv.NewMat()
	if ok := s.Video.Read(&currentFrame); !ok || currentFrame.Empty() {
		return nil
	}

	frame, err := frame.NewFrame(frameIndex, currentFrame)
	if err != nil {
		return nil
	}

	return frame
}
