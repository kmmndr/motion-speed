package motion

import (
	"log"

	"motionspeed/internal/frame"
	"motionspeed/internal/video"
)

type Input struct {
	stream           *video.Stream
	threshold        float64
	cameraViewLength float64
	frameBuffer      *frame.FrameBuffer
}

func NewInput(stream *video.Stream, threshold float64, length float64) *Input {
	return &Input{
		stream:           stream,
		threshold:        threshold,
		cameraViewLength: length,
		frameBuffer:      frame.NewFrameBuffer(),
	}
}

func (i *Input) TimeAtFrame(frame *frame.Frame) float64 {
	return i.stream.TimeAtFrame(frame)
}

func (i *Input) Fps() float64 {
	return i.stream.Fps()
}

func (i *Input) Speed(duration float64) float64 {
	return (i.cameraViewLength / duration) * 3.6
}

func (i *Input) DetectMotion(afterMotion func(*Motion)) {
	i.detect(func(*frame.Frame) {}, func(*frame.Frame) {}, afterMotion)
}

func (i *Input) Detect(onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	i.detect(onMotionStart, onMotionEnd, afterMotion)
}

func (i *Input) detect(onMotionStart func(*frame.Frame), onMotionEnd func(*frame.Frame), afterMotion func(*Motion)) {
	var currentFrame *frame.Frame
	var startFrame *frame.Frame
	var endFrame *frame.Frame

	frameIndex := 0
	isMotionDetected := false

	for {
		if currentFrame = i.stream.Read(frameIndex); currentFrame == nil {
			break
		}

		grayFrame, err := currentFrame.Gray()
		if err != nil {
			log.Printf("Unable to create Frame : %v", err)
		}

		frameIndex++

		i.frameBuffer.UpdateAverageFrame(grayFrame)
		diffPercentage := i.frameBuffer.DiffPercentage(grayFrame)

		if diffPercentage < 0 {
			log.Fatal("ko")
		}

		if diffPercentage > i.threshold {
			i.frameBuffer.UpdateAverageDiffPercentage(grayFrame)

			if !isMotionDetected { // Motion start
				isMotionDetected = true
				startFrame, _ = currentFrame.Clone()

				onMotionStart(startFrame)
			}
		} else if isMotionDetected { // Motion end
			isMotionDetected = false
			endFrame, _ = currentFrame.Clone()

			onMotionEnd(endFrame)

			if startFrame == nil {
				log.Fatalf("err : %v", err)
			}
			motion, err := NewMotion(startFrame, endFrame, i.frameBuffer.MeanDiffPercentage())
			if err != nil {
				log.Fatalf("err : %v", err)
			}
			afterMotion(motion)

			i.frameBuffer.Reset()

			startFrame.Close()
			endFrame.Close()
		}

		grayFrame.Close()
		currentFrame.Close()
	}
}
