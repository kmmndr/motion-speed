package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"motionspeed/internal/motion"
	"motionspeed/internal/video"
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

	video, err := video.OpenVideo(videoPath)
	if err != nil {
		log.Fatalf("Error: unable to open video file: %v\n", err)
	}
	defer video.Close()

	motionDetector := motion.NewMotionDetector(movementThreshold, cameraViewLength)
	motionDetector.Detect(video)
}
