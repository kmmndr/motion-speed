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

var config Config

type Config struct {
	videoPath string
	videoUrl  string

	motionThreshold  float64
	cameraViewLength float64

	printMotion bool
	printTmpl   string

	commandTmpl string

	saveFrames bool
}

func init() {
	config = Config{}

	flag.StringVar(&config.videoPath, "video-file", "", "Video file")
	flag.StringVar(&config.videoUrl, "video-url", "", "Video url")

	flag.Float64Var(&config.motionThreshold, "motion-threshold", 0.5, "Motion threshold %")
	flag.Float64Var(&config.cameraViewLength, "length", 0.0, "Length of the camera view")

	flag.BoolVar(&config.printMotion, "print", false, "print motion")
	flag.StringVar(&config.printTmpl, "print-format", "", "print format")

	flag.StringVar(&config.commandTmpl, "command", "", "command line to run after motion (ex: echo '{{.Date}} {{.Duration}} {{.Speed}})'")

	flag.BoolVar(&config.saveFrames, "save-frames", false, "Save start/end frames")

	flag.Parse()
}

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

	if config.videoPath == "" && config.videoUrl == "" {
		fmt.Println("Error: missing video file or url option")
		os.Exit(1)
	}

	if config.printTmpl == "" {
		config.printTmpl = "Date: {{.Date}}\n" +
			"Motion duration: {{.Duration}} seconds\n" +
			"Mean Diff Percentage: {{.MeanDiffPercentage}}%\n" +
			"Speed: {{.Speed}} km/h\n"
	}

	var stream *video.Stream
	var err error

	if config.videoUrl != "" {
		stream, err = video.NewDeviceStream(config.videoUrl)
		if err != nil {
			log.Fatalf("Error: unable to open video url: %v\n", err)
		}
		defer stream.Close()
	} else if config.videoPath != "" {
		stream, err = video.NewFileStream(config.videoPath)
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

	sensor := motion.NewSensor(stream, config.motionThreshold, config.cameraViewLength)
	sensor.Detect(
		func(startFrame *frame.Frame) {
			if config.saveFrames {
				startTime := sensor.TimeAtFrame(startFrame)
				fmt.Printf("Motion started at: %.2f seconds.\n", startTime)
				if !gocv.IMWrite("motion-start.jpg", *startFrame.Mat()) {
					log.Fatal("Unable to write image")
				}
			}
		},
		func(endFrame *frame.Frame) {
			if config.saveFrames {
				endTime := sensor.TimeAtFrame(endFrame)
				fmt.Printf("Motion ended at: %.2f seconds.\n", endTime)
				if !gocv.IMWrite("motion-end.jpg", *endFrame.Mat()) {
					log.Fatal("Unable to write image")
				}
			}
		},
		func(detectedMotion *motion.Motion) {
			motionReport := motion.NewMotionReport(detectedMotion, sensor)

			if config.printMotion {
				str := expandTemplate(config.printTmpl, motionReport)

				fmt.Printf("%s\n", str)
			}

			if config.commandTmpl != "" {
				str := expandTemplate(config.commandTmpl, motionReport)

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
