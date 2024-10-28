package motion

import (
	"fmt"
	"time"
)

type MotionReport struct {
	motion *Motion
	sensor *Sensor

	UUID               string `json:"uuid"`
	Duration           string `json:"duration"`
	Speed              string `json:"speed"`
	Date               string `json:"date"`
	MeanDiffPercentage string `json:"mean_diff_percentage"`
}

func NewMotionReport(motion *Motion, sensor *Sensor) *MotionReport {
	motionDuration := float64(motion.FramesCount()) / float64(sensor.Fps())
	speed := (sensor.cameraViewLength / motionDuration) * 3.6
	now := time.Now().Format(time.RFC3339)

	return &MotionReport{
		motion: motion,
		sensor: sensor,

		UUID:               motion.UUID(),
		Duration:           fmt.Sprintf("%.2f", motionDuration),
		Speed:              fmt.Sprintf("%.2f", speed),
		Date:               now,
		MeanDiffPercentage: fmt.Sprintf("%.2f", motion.MeanDiffPercentage()),
	}
}
