package frame

import (
	"gocv.io/x/gocv"
)

type FrameBuffer struct {
	frameIndex            int
	averageMat            *gocv.Mat
	lastDiffPercentage    float64
	averageDiffPercentage float64
}

func NewFrameBuffer() *FrameBuffer {
	averageMat := gocv.NewMat()

	return &FrameBuffer{
		frameIndex:            0,
		averageMat:            &averageMat,
		lastDiffPercentage:    0,
		averageDiffPercentage: -1,
	}
}

func (fb *FrameBuffer) FramesCountSinceStart(currentFrame *Frame) int {
	return currentFrame.FrameIndex() - fb.frameIndex
}

func (fb *FrameBuffer) UpdateAverageFrame(currentFrame *Frame) {
	if fb.averageMat.Empty() {
		currentFrame.mat.ConvertTo(fb.averageMat, gocv.MatTypeCV64F)
	} else {
		gocv.AccumulatedWeighted(*currentFrame.mat, fb.averageMat, 0.5)
	}
}

func (fb *FrameBuffer) Reset() {
	fb.averageDiffPercentage = -1
	fb.lastDiffPercentage = 0
}

func (fb *FrameBuffer) UpdateAverageDiffPercentage(currentFrame *Frame) {
	if fb.averageDiffPercentage < 0 {
		fb.frameIndex = currentFrame.FrameIndex()
		fb.averageDiffPercentage = fb.DiffPercentage(currentFrame)
	} else {
		fb.averageDiffPercentage =
			fb.averageDiffPercentage +
				(fb.DiffPercentage(currentFrame)-fb.averageDiffPercentage)/float64(fb.FramesCountSinceStart(currentFrame))
	}
	fb.lastDiffPercentage = fb.DiffPercentage(currentFrame)
}

func (fb *FrameBuffer) LastDiffPercentage() float64 {
	return fb.lastDiffPercentage
}
func (fb *FrameBuffer) MeanDiffPercentage() float64 {
	return fb.averageDiffPercentage
}

func (fb *FrameBuffer) DiffPercentage(currentFrame *Frame) float64 {
	return float64(fb.PixelsDiffCount(currentFrame)) * 100.0 / float64(currentFrame.Pixels())
}

func (fb *FrameBuffer) PixelsDiffCount(currentFrame *Frame) int {
	averageConverted := gocv.NewMat()
	defer averageConverted.Close()

	diff := gocv.NewMat()
	defer diff.Close()

	fb.averageMat.ConvertTo(&averageConverted, gocv.MatTypeCV8U)

	gocv.AbsDiff(*currentFrame.mat, averageConverted, &diff)
	gocv.Threshold(diff, &diff, 25, 255, gocv.ThresholdBinary)
	nonZeroPixels := gocv.CountNonZero(diff)

	return nonZeroPixels
}
