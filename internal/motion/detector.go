package motion

import (
	"fmt"
	"log"
	"motionspeed/internal/frame"

	"gocv.io/x/gocv"
)

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

func (md *MotionDetector) Detect(video *gocv.VideoCapture) {
	fps := video.Get(gocv.VideoCaptureFPS)
	if fps <= 0 {
		log.Fatal("Error: unable to get video frame rate")
	}
	fmt.Printf("Video frame rate: %.2f fps\n", fps)

	frameIndex := 0
	isMovementDetected := false
	movementFrameCount := 0

	for {
		currentFrame := gocv.NewMat()
		if ok := video.Read(&currentFrame); !ok || currentFrame.Empty() {
			break
		}

		frame, err := frame.NewFrame(frameIndex, currentFrame)
		if err != nil {
			log.Printf("Unable to create Frame : %v", err)
		}
		grayFrame, err := frame.Gray()
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

				movementTime := float64(frameIndex) / fps
				fmt.Printf("Motion starting at: %.2f seconds.\n", movementTime)

				if !gocv.IMWrite("motion-start.jpg", *grayFrame.Mat()) {
					log.Fatal("Unable to write image")
				}
			}
			movementFrameCount++
		} else if isMovementDetected {
			isMovementDetected = false

			if !gocv.IMWrite("motion-end.jpg", *grayFrame.Mat()) {
				log.Fatal("Unable to write image")
			}
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
