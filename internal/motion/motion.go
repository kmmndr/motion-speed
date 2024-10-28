package motion

import (
	"errors"
	"log"

	uuid "github.com/gofrs/uuid/v5"

	"motionspeed/internal/frame"
)

type Motion struct {
	startFrame         *frame.Frame
	endFrame           *frame.Frame
	meanDiffPercentage float64
	uuid               uuid.UUID
}

func NewMotion(startFrame *frame.Frame, endFrame *frame.Frame, meanDiffPercentage float64) (*Motion, error) {
	if startFrame.Mat().Closed() || endFrame.Mat().Closed() {
		return nil, errors.New("Frame is empty")
	}

	ref, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
	}

	return &Motion{
		startFrame:         startFrame,
		endFrame:           endFrame,
		meanDiffPercentage: meanDiffPercentage,
		uuid:               ref,
	}, nil
}

func (m *Motion) MeanDiffPercentage() float64 {
	return m.meanDiffPercentage
}

func (m *Motion) UUID() string {
	return m.uuid.String()
}

func (m *Motion) FramesCount() int {
	return (m.endFrame.FrameIndex() - m.startFrame.FrameIndex())
}
