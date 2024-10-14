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
	var videoUrl string
	var motionThreshold float64
	var cameraViewLength float64
	var printMotion bool
	var commandTmpl string

	flag.Float64Var(&motionThreshold, "motion-threshold", 0.5, "Motion threshold %")
	flag.Float64Var(&cameraViewLength, "length", 0.0, "Length of the camera view")
	flag.StringVar(&videoPath, "video-file", "", "Video file")
	flag.StringVar(&videoUrl, "video-url", "", "Video url")
	flag.BoolVar(&printMotion, "print", false, "print motion")
	flag.StringVar(&commandTmpl, "command", "", "command line to run after motion (ex: echo {{.Date}} {{.Duration}} {{.Speed}})")
	flag.Parse()

	if videoPath == "" && videoUrl == "" {
		fmt.Println("Error: missing video file or url option")
		os.Exit(1)
	}

	var stream *video.Stream
	var err error

	if videoUrl != "" {
		stream, err = video.NewDeviceStream(videoUrl)
		if err != nil {
			log.Fatalf("Error: unable to open video url: %v\n", err)
		}
		defer stream.Close()
	} else if videoPath != "" {
		stream, err = video.NewFileStream(videoPath)
		if err != nil {
			log.Fatalf("Error: unable to open video file: %v\n", err)
		}
		defer stream.Close()
	}

	fps := stream.Fps()
	if fps <= 0 {
		log.Fatal("Error: unable to get video frame rate")
	}
	fmt.Printf("Video frame rate: %.2f fps\n", fps)

	motionDetector := motion.NewMotionDetector(motionThreshold, cameraViewLength)
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
			motionDuration := float64(detectedMotion.FramesCount()) / fps
			speed := (cameraViewLength / motionDuration) * 3.6
			now := time.Now().Format(time.RFC3339)

			if printMotion {
				fmt.Printf("Motion duration: %.2f seconds.\n", motionDuration)
				fmt.Printf("Mean Diff Percentage: %.2f%%.\n", detectedMotion.MeanDiffPercentage())
				fmt.Printf("Speed: %.2f km/h.\n", speed)
				fmt.Printf("Date: %s\n\n", now)
			}

			if commandTmpl != "" {
				t, err := template.New("greeting").Parse(commandTmpl)
				if err != nil {
					log.Fatalf("Error parsing template: %v", err)
				}

				data := struct {
					Duration string
					Speed    string
					Date     string
				}{
					Duration: fmt.Sprintf("%.2f", motionDuration),
					Speed:    fmt.Sprintf("%.2f", speed),
					Date:     now,
				}
				var buf bytes.Buffer

				if err := t.Execute(&buf, data); err != nil {
					log.Fatalf("Error executing template: %v", err)
				}

				fmt.Printf("Executing command: %s\n", buf.String())

				cmd := exec.Command("sh", "-c", buf.String())

				var out bytes.Buffer
				cmd.Stdout = &out

				if err := cmd.Run(); err != nil {
					log.Fatalf("Error executing command: %v", err)
				}

				fmt.Println(out.String())
			}
		})
}
