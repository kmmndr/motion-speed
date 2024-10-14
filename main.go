package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"

	"motionspeed/internal/frame"
	"motionspeed/internal/motion"
	"motionspeed/internal/video"

	"gocv.io/x/gocv"
)

func expandTemplate(tmpl string, report *motion.MotionReport) string {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, report); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	return buf.String()
}

func main() {
	var videoPath string
	var videoUrl string
	var motionThreshold float64
	var cameraViewLength float64
	var printMotion bool
	var printTmpl string
	var commandTmpl string

	flag.Float64Var(&motionThreshold, "motion-threshold", 0.5, "Motion threshold %")
	flag.Float64Var(&cameraViewLength, "length", 0.0, "Length of the camera view")
	flag.StringVar(&videoPath, "video-file", "", "Video file")
	flag.StringVar(&videoUrl, "video-url", "", "Video url")
	flag.BoolVar(&printMotion, "print", false, "print motion")
	flag.StringVar(&printTmpl, "print-format", "", "print format")
	flag.StringVar(&commandTmpl, "command", "", "command line to run after motion (ex: echo {{.Date}} {{.Duration}} {{.Speed}})")
	flag.Parse()

	if videoPath == "" && videoUrl == "" {
		fmt.Println("Error: missing video file or url option")
		os.Exit(1)
	}

	if printTmpl == "" {
		printTmpl = "Date: {{.Date}}\n" +
			"Motion duration: {{.Duration}} seconds\n" +
			"Mean Diff Percentage: {{.MeanDiffPercentage}}%\n" +
			"Speed: {{.Speed}} km/h\n"
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

	sensor := motion.NewSensor(stream, motionThreshold, cameraViewLength)
	sensor.Detect(
		func(startFrame *frame.Frame) {
			startTime := sensor.TimeAtFrame(startFrame)
			fmt.Printf("Motion started at: %.2f seconds.\n", startTime)
			if !gocv.IMWrite("motion-start.jpg", *startFrame.Mat()) {
				log.Fatal("Unable to write image")
			}
		},
		func(endFrame *frame.Frame) {
			endTime := sensor.TimeAtFrame(endFrame)
			fmt.Printf("Motion ended at: %.2f seconds.\n", endTime)
			if !gocv.IMWrite("motion-end.jpg", *endFrame.Mat()) {
				log.Fatal("Unable to write image")
			}
		},
		func(detectedMotion *motion.Motion) {
			motionReport := motion.NewMotionReport(detectedMotion, sensor)

			if printMotion {
				str := expandTemplate(printTmpl, motionReport)

				fmt.Printf("%s\n", str)
			}

			if commandTmpl != "" {
				str := expandTemplate(commandTmpl, motionReport)

				fmt.Printf("Executing command: %s\n", str)

				cmd := exec.Command("sh", "-c", str)

				var out bytes.Buffer
				cmd.Stdout = &out

				if err := cmd.Run(); err != nil {
					log.Fatalf("Error executing command: %v", err)
				}

				fmt.Println(out.String())
			}
		})
}
