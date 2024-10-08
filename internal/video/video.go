package video

import (
	"fmt"

	"gocv.io/x/gocv"
)

func OpenVideo(videoPath string) (*gocv.VideoCapture, error) {
	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open video file: %v", err)
	}
	return video, nil
}
