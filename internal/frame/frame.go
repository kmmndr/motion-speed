package frame

import (
	"errors"

	"gocv.io/x/gocv"
)

type Frame struct {
	frameIndex int
	mat        *gocv.Mat
}

func NewFrame(frameIndex int, mat *gocv.Mat) (*Frame, error) {
	if mat.Empty() {
		return nil, errors.New("Frame is empty")
	}

	return &Frame{frameIndex: frameIndex, mat: mat}, nil
}

func (f *Frame) Mat() *gocv.Mat {
	return f.mat
}

func (f *Frame) FrameIndex() int {
	return f.frameIndex
}

func (f *Frame) Gray() (*Frame, error) {
	gray := gocv.NewMat()
	gocv.CvtColor(*f.mat, &gray, gocv.ColorBGRToGray)

	return NewFrame(f.frameIndex, &gray)
}

func (f *Frame) Clone() (*Frame, error) {
	clone := f.mat.Clone()

	return NewFrame(f.frameIndex, &clone)
}

func (f *Frame) Height() int {
	return f.mat.Rows()
}

func (f *Frame) Width() int {
	return f.mat.Cols()
}

func (f *Frame) Pixels() int {
	return f.Height() * f.Width()
}

func (f *Frame) Close() {
	f.mat.Close()
}
