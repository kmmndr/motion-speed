package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"text/template"
	"time"

	"motionspeed/internal/frame"
	"motionspeed/internal/motion"
	"motionspeed/internal/video"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gocv.io/x/gocv"
)

var logger *slog.Logger
var config Config

type Config struct {
	videoPath string
	videoUrl  string

	motionThreshold  float64
	cameraViewLength float64

	printMotion bool
	printJson   bool
	printTmpl   string

	commandTmpl string

	saveFrames bool

	mqtt                bool
	mqttBroker          string
	mqttPort            int
	mqttClientId        string
	mqttCaFile          string
	mqttCertificateFile string
	mqttKeyFile         string
}

func init() {
	var mqttPortString string

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config = Config{}

	flag.StringVar(&config.videoPath, "video-file", "", "Video file")
	flag.StringVar(&config.videoUrl, "video-url", "", "Video url")

	flag.Float64Var(&config.motionThreshold, "motion-threshold", 0.5, "Motion threshold %")
	flag.Float64Var(&config.cameraViewLength, "length", 0.0, "Length of the camera view")

	flag.BoolVar(&config.printMotion, "print", false, "print motion")
	flag.BoolVar(&config.printJson, "print-json", false, "print motion as json")
	flag.StringVar(&config.printTmpl, "print-format", "", "print format")

	flag.StringVar(&config.commandTmpl, "command", "", "command line to run after motion (ex: echo '{{.Date}} {{.Duration}} {{.Speed}})'")

	flag.BoolVar(&config.saveFrames, "save-frames", false, "Save start/end frames")

	flag.BoolVar(&config.mqtt, "mqtt", false, "Enable MQTT")
	flag.StringVar(&config.mqttBroker, "mqtt-broker", "", "MQTT Broker")
	flag.StringVar(&mqttPortString, "mqtt-port", "", "MQTT Port")
	flag.StringVar(&config.mqttClientId, "mqtt-client-id", "motion-speed", "MQTT Client ID")
	flag.StringVar(&config.mqttCaFile, "mqtt-ca-file", "", "MQTT CA file")
	flag.StringVar(&config.mqttCertificateFile, "mqtt-certificate-file", "", "MQTT Certificate file")
	flag.StringVar(&config.mqttKeyFile, "mqtt-key-file", "", "MQTT Key file")

	flag.Parse()

	if port, err := strconv.Atoi(mqttPortString); err == nil {
		config.mqttPort = port
	}
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

func mqttClientOptions() *mqtt.ClientOptions {
	certpool := x509.NewCertPool()

	if config.mqttCaFile != "" {
		ca, err := os.ReadFile(config.mqttCaFile)
		if err != nil {
			log.Fatalln(err.Error())
		}
		certpool.AppendCertsFromPEM(ca)
	}

	clientKeyPair, err := tls.LoadX509KeyPair(config.mqttCertificateFile, config.mqttKeyFile)
	if err != nil {
		panic(err)
	}

	tlsConfig := tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}

	opts := mqtt.NewClientOptions()

	opts.AddBroker(fmt.Sprintf("mqtts://%s:%d", config.mqttBroker, config.mqttPort))
	opts.SetClientID(config.mqttClientId)
	opts.SetTLSConfig(&tlsConfig)
	opts.OnConnect = func(client mqtt.Client) {
		logger.Info("Connected to MQTT server")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Warn(fmt.Sprintf("Connect lost: %v", err))
	}
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	return opts
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

	logger.Info("Start your engine \\o//")

	var stream *video.Stream
	var mqttClient mqtt.Client
	var err error

	if config.mqtt {
		mqttClient = mqtt.NewClient(mqttClientOptions())
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		defer mqttClient.Disconnect(250)
	}

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
	logger.Debug(fmt.Sprintf("Video frame rate: %.2f fps\n", fps))

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
				var str string

				if config.printJson {
					motionReportJson, _ := json.Marshal(motionReport)
					str = string(motionReportJson)
				} else {
					str = expandTemplate(config.printTmpl, motionReport)
				}

				fmt.Printf("%s\n", str)
			}

			if config.mqtt {
				motionReportJson, _ := json.Marshal(motionReport)

				text := string(motionReportJson)
				logger.Info(fmt.Sprintf("Publishing MQTT Message: %s", text))
				token := mqttClient.Publish("motions", 0, false, text)
				token.Wait()
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
