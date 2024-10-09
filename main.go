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

	motionDetector := motion.NewMotionDetector(movementThreshold, cameraViewLength)
	motionDetector.Detect(stream.Video,
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
		})
}
