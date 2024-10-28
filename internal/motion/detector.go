package motion

import (
	"errors"
	"fmt"
	"log"
	"time"

	uuid "github.com/gofrs/uuid/v5"

	"motionspeed/internal/frame"
	"motionspeed/internal/video"
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

type Sensor struct {
	stream           *video.Stream
	threshold        float64
	cameraViewLength float64
	frameBuffer      *frame.FrameBuffer
}

func NewSensor(stream *video.Stream, threshold float64, length float64) *Sensor {
	return &Sensor{
		stream:           stream,
		threshold:        threshold,
		cameraViewLength: length,
		frameBuffer:      frame.NewFrameBuffer(),
	}
}

func (s *Sensor) TimeAtFrame(frame *frame.Frame) float64 {
	return s.stream.TimeAtFrame(frame)
}

func (s *Sensor) Fps() float64 {
	return s.stream.Fps()
}

func (s *Sensor) Speed(duration float64) float64 {
	return (s.cameraViewLength / duration) * 3.6
}

func (s *Sensor) Detect(onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	s.detect(onMotionStart, onMotionEnd, afterMotion)
}

func (s *Sensor) detect(onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	var currentFrame *frame.Frame
	var startFrame *frame.Frame
	var endFrame *frame.Frame

	frameIndex := 0
	isMovementDetected := false
	movementFrameCount := 0

	for {
		if currentFrame = s.stream.Read(frameIndex); currentFrame == nil {
			break
		}

		grayFrame, err := currentFrame.Gray()
		if err != nil {
			log.Printf("Unable to create Frame : %v", err)
		}
		defer grayFrame.Close()

		frameIndex++

		s.frameBuffer.UpdateAverageFrame(grayFrame)
		diffPercentage := s.frameBuffer.DiffPercentage(grayFrame)

		if diffPercentage < 0 {
			log.Fatal("ko")
		}

		if diffPercentage > s.threshold {
			s.frameBuffer.UpdateAverageDiffPercentage(grayFrame)

			if !isMovementDetected { // Motion start
				isMovementDetected = true
				startFrame, _ = currentFrame.Clone()

				onMotionStart(startFrame)
			}
			movementFrameCount++
		} else if isMovementDetected { // Motion end
			isMovementDetected = false
			endFrame, _ = currentFrame.Clone()

			onMotionEnd(endFrame)

			if startFrame == nil {
				log.Fatalf("err : %v", err)
			}
			motion, err := NewMotion(startFrame, endFrame, s.frameBuffer.MeanDiffPercentage())
			if err != nil {
				log.Fatalf("err : %v", err)
			}
			afterMotion(motion)

			s.frameBuffer.Reset()

			startFrame.Close()
			startFrame = nil
			endFrame.Close()
			endFrame = nil
		}

		currentFrame.Close()
		currentFrame = nil
	}
}
