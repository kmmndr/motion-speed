package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gocv.io/x/gocv"
)

const (
	frameWindowSize = 2
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

func (f *Frame) Gray() *Frame {
	gray := gocv.NewMat()

	if f.mat.Empty() {
		log.Fatal("Frame is empty")
	}

	gocv.CvtColor(f.mat, &gray, gocv.ColorBGRToGray)

	return NewFrame(f.number, gray)
}

func (f *Frame) Clone() *Frame {
	return NewFrame(f.number, f.mat.Clone())
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

func main() {
	var videoPath string
	var movementThreshold int
	var cameraViewLength float64

	flag.IntVar(&movementThreshold, "threshold", 2000, "Movement threshold")
	flag.Float64Var(&cameraViewLength, "length", 0.0, "Length of the camera view")
	flag.StringVar(&videoPath, "video", "", "Video filename")
	flag.Parse()

	if videoPath == "" {
		fmt.Println("Error: missing video filename option")
		os.Exit(1)
	}

	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		fmt.Printf("Error: unable to open video file: %v\n", err)
		return
	}
	defer video.Close()

	fps := video.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		fmt.Println("Error: unable to get video frame rate")
		return
	}
	fmt.Printf("Video frame rate: %.2f fps\n", fps)

	frames := NewFrameBuffer()
	// defer frames.Close()

	frameIndex := 0
	isMovementDetected := false
	movementFrameCount := 0

	for {
		currentFrame := gocv.NewMat()
		if ok := video.Read(&currentFrame); !ok || currentFrame.Empty() {
			break
		}

		frame := NewFrame(frameIndex, currentFrame)
		grayFrame := frame.Gray()
		frame.Close()
		defer grayFrame.Close()

		frameIndex++

		nonZeroPixels := frames.PixelsDiffCount(grayFrame)

		if nonZeroPixels < 0 {
			log.Fatal("ko")
		}

		if nonZeroPixels > movementThreshold {
			if !isMovementDetected {
				isMovementDetected = true

				movementTime := float64(frameIndex) / fps
				fmt.Printf("Motion starting at: %.2f seconds.\n", movementTime)

				if !gocv.IMWrite("motion-start.jpg", grayFrame.mat) {
					log.Fatal("Unable to write image")
				}
			}

			movementFrameCount++
		} else if isMovementDetected {
			isMovementDetected = false

			if !gocv.IMWrite("motion-end.jpg", grayFrame.mat) {
				log.Fatal("Unable to write image")
			}
		}
	}

	if movementFrameCount > 0 {
		movementTime := float64(movementFrameCount) / fps
		fmt.Printf("Motion duration: %.2f seconds.\n", movementTime)
		movementSpeed := (cameraViewLength / movementTime) * 3.6
		fmt.Printf("Speed: %.2f km/h.\n", movementSpeed)
	} else {
		fmt.Println("Motion not detected")
	}
}
