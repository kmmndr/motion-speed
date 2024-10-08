package frame

import (
	"log"

	"gocv.io/x/gocv"
)

type Frame struct {
	number int
	mat    gocv.Mat
}

func NewFrame(number int, mat gocv.Mat) *Frame {
	if mat.Empty() {
		log.Fatal("Frame is empty")
	}
	return &Frame{number: number, mat: mat}
}

func (f *Frame) Mat() *gocv.Mat {
	return &f.mat
}

func (f *Frame) Gray() *Frame {
	gray := gocv.NewMat()
	gocv.CvtColor(f.mat, &gray, gocv.ColorBGRToGray)
	return NewFrame(f.number, gray)
}

func (f *Frame) Close() {
	f.mat.Close()
}

type FrameBuffer struct {
	average gocv.Mat
}

func NewFrameBuffer() *FrameBuffer {
	return &FrameBuffer{
		average: gocv.NewMat(),
	}
}

func (fb *FrameBuffer) PixelsDiffCount(currentFrame *Frame) int {
	if fb.average.Empty() {
		currentFrame.mat.ConvertTo(&fb.average, gocv.MatTypeCV64F)
	} else {
		gocv.AccumulatedWeighted(currentFrame.mat, &fb.average, 0.5)
	}

	averageConverted := gocv.NewMat()
	defer averageConverted.Close()

	diff := gocv.NewMat()
	defer diff.Close()

	fb.average.ConvertTo(&averageConverted, gocv.MatTypeCV8U)

	gocv.AbsDiff(currentFrame.mat, averageConverted, &diff)
	gocv.Threshold(diff, &diff, 25, 255, gocv.ThresholdBinary)
	nonZeroPixels := gocv.CountNonZero(diff)

	return nonZeroPixels
}
