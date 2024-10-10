package motion

import (
	"fmt"
	"log"
	"motionspeed/internal/frame"
	"motionspeed/internal/video"
)

type Motion struct {
	startFrame *frame.Frame
	endFrame   *frame.Frame
}

func NewMotion(startFrame *frame.Frame, endFrame *frame.Frame) (*Motion, error) {
	if startFrame.Mat().Closed() || endFrame.Mat().Closed() {
		return nil, errors.New("Frame is empty")
	}

	return &Motion{
		startFrame: startFrame,
		endFrame:   endFrame,
	}, nil
}

func (m *Motion) FramesCount() int {
	return (m.endFrame.FrameIndex() - m.startFrame.FrameIndex())
}

type MotionDetector struct {
	threshold        int
	cameraViewLength float64
	frameBuffer      *frame.FrameBuffer
}

func NewMotionDetector(threshold int, length float64) *MotionDetector {
	return &MotionDetector{
		threshold:        threshold,
		cameraViewLength: length,
		frameBuffer:      frame.NewFrameBuffer(),
	}
}

func (md *MotionDetector) Speed(duration float64) float64 {
	return (md.cameraViewLength / duration) * 3.6
}

func (md *MotionDetector) Detect(stream *video.Stream, onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	md.detect(stream, onMotionStart, onMotionEnd, afterMotion)
}

func (md *MotionDetector) detect(stream *video.Stream, onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	var currentFrame *frame.Frame
	var startFrame *frame.Frame
	var endFrame *frame.Frame

	frameIndex := 0
	isMovementDetected := false
	movementFrameCount := 0

	for {
		if currentFrame = stream.Read(frameIndex); currentFrame == nil {
			break
		}

		grayFrame, err := currentFrame.Gray()
		if err != nil {
			log.Printf("Unable to create Frame : %v", err)
		}
		defer grayFrame.Close()

		frameIndex++

		nonZeroPixels := md.frameBuffer.PixelsDiffCount(grayFrame)

		if nonZeroPixels < 0 {
			log.Fatal("ko")
		}

		if nonZeroPixels > md.threshold {
			if !isMovementDetected {
				isMovementDetected = true
				onMotionStart(frame)
			}
			movementFrameCount++
		} else if isMovementDetected {
			isMovementDetected = false
			onMotionEnd(frame)
		}

		frame.Close()
	}

	if movementFrameCount > 0 {
		movementTime := float64(movementFrameCount) / fps
		fmt.Printf("Motion duration: %.2f seconds.\n", movementTime)
		speed := (md.cameraViewLength / movementTime) * 3.6
		fmt.Printf("Speed: %.2f km/h.\n", speed)
	} else {
		fmt.Println("Motion not detected")
	}
}
