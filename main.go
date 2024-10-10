package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"motionspeed/internal/frame"
	"motionspeed/internal/motion"
	"motionspeed/internal/video"

	"gocv.io/x/gocv"
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

	stream, err := video.NewStream(videoPath)
	if err != nil {
		log.Fatalf("Error: unable to open video file: %v\n", err)
	}
	defer stream.Close()

	fps := stream.Fps()
	if fps <= 0 {
		log.Fatal("Error: unable to get video frame rate")
	}
	fmt.Printf("Video frame rate: %.2f fps\n", fps)

	motionDetector := motion.NewMotionDetector(movementThreshold, cameraViewLength)
	motionDetector.Detect(stream,
		func(startFrame *frame.Frame) {
			startTime := stream.TimeAtFrame(startFrame)
			fmt.Printf("Motion started at: %.2f seconds.\n", startTime)
			if !gocv.IMWrite("motion-start.jpg", *startFrame.Mat()) {
				log.Fatal("Unable to write image")
			}
		},
		func(endFrame *frame.Frame) {
			endTime := stream.TimeAtFrame(endFrame)
			fmt.Printf("Motion ended at: %.2f seconds.\n", endTime)
			if !gocv.IMWrite("motion-end.jpg", *endFrame.Mat()) {
				log.Fatal("Unable to write image")
			}
		},
		func(detectedMotion *motion.Motion) {
			movementTime := float64(detectedMotion.FramesCount()) / fps
			fmt.Printf("Motion duration: %.2f seconds.\n", movementTime)
			speed := (cameraViewLength / movementTime) * 3.6
			fmt.Printf("Speed: %.2f km/h.\n", speed)
		})
}
