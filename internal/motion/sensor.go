package motion

import (
	"log"

	"motionspeed/internal/frame"
	"motionspeed/internal/video"
)

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

func (s *Sensor) DetectMotion(afterMotion func(*Motion)) {
	s.detect(func(*frame.Frame) {}, func(*frame.Frame) {}, afterMotion)
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
