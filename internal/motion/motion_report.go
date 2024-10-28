package motion

import (
	"fmt"
	"time"
)

type MotionReport struct {
	motion *Motion
	input  *Input

	UUID               string `json:"uuid"`
	Duration           string `json:"duration"`
	Speed              string `json:"speed"`
	Date               string `json:"date"`
	MeanDiffPercentage string `json:"mean_diff_percentage"`
}

func NewMotionReport(motion *Motion, input *Input) *MotionReport {
	motionDuration := float64(motion.FramesCount()) / float64(input.Fps())
	speed := (input.cameraViewLength / motionDuration) * 3.6
	now := time.Now().Format(time.RFC3339)

	return &MotionReport{
		motion: motion,
		input:  input,

		UUID:               motion.UUID(),
		Duration:           fmt.Sprintf("%.2f", motionDuration),
		Speed:              fmt.Sprintf("%.2f", speed),
		Date:               now,
		MeanDiffPercentage: fmt.Sprintf("%.2f", motion.MeanDiffPercentage()),
	}
}
