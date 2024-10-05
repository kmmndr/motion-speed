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

	frame := gocv.NewMat()
	defer frame.Close()

	gray := gocv.NewMat()
	defer gray.Close()

	average := gocv.NewMat()
	defer average.Close()

	diff := gocv.NewMat()
	defer diff.Close()

	frameIndex := 0
	isMovementDetected := false
	movementFrameCount := 0

	frameBuffer := make([]gocv.Mat, 0, frameWindowSize)

	for {
		if ok := video.Read(&frame); !ok || frame.Empty() {
			break
		}
		frameIndex++

		gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)

		if average.Empty() {
			average = gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV64F)
		}

		if len(frameBuffer) >= frameWindowSize {
			frameBuffer[0].Close()
			frameBuffer = frameBuffer[1:]
		}
		frameBuffer = append(frameBuffer, gray.Clone())

		if len(frameBuffer) == frameWindowSize {
			average.SetTo(gocv.NewScalar(0.0, 0.0, 0.0, 0.0))

			for _, f := range frameBuffer {
				temp := gocv.NewMat()
				f.ConvertTo(&temp, gocv.MatTypeCV64F)
				gocv.AddWeighted(average, 1.0, temp, 1.0/float64(frameWindowSize), 0.0, &average)
				temp.Close()
			}

			averageConverted := gocv.NewMat()
			defer averageConverted.Close()
			average.ConvertTo(&averageConverted, gocv.MatTypeCV8U)

			gocv.AbsDiff(gray, averageConverted, &diff)

			gocv.Threshold(diff, &diff, 25, 255, gocv.ThresholdBinary)

			nonZeroPixels := gocv.CountNonZero(diff)

			if nonZeroPixels > movementThreshold {
				if !isMovementDetected {
					isMovementDetected = true

					movementTime := float64(frameIndex) / fps
					fmt.Printf("Motion starting at: %.2f seconds.\n", movementTime)

					if !gocv.IMWrite("motion-start.jpg", frame) {
						log.Fatal("Unable to write image")
					}
				}

				movementFrameCount++
			} else if isMovementDetected {
				isMovementDetected = false

				if !gocv.IMWrite("motion-end.jpg", frame) {
					log.Fatal("Unable to write image")
				}
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

	for _, f := range frameBuffer {
		f.Close()
	}
}
