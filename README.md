# Motion Speed

motion-speed is a tool designed to calculate the speed of objects by
measuring the time taken to traverse a known distance within the field of view.
To use this tool, measure the length (in meters) of the path visible by the
camera. Pass this distance using the --length parameter. The tool
then calculates speed by recording the time taken by the object to cross this
path.

## Installation

### Using Golang

motion-speed uses [GoCV](https://gocv.io/) and needs
[OpenCV](https://opencv.org/) librairies to compile.

```
# Installation example on AlpineLinux
apk add --no-cache go build-base opencv-dev \
  opencv libopencv_aruco libopencv_photo libopencv_video
```

```
go install github.com/kmmndr/motion-speed@latest
```

### Using Docker

```
docker build -t motion-speed .
```

## Usage

```
# Basic usage to print recorded data in console
motion-speed -video-url 'rtsp://user:password@127.0.0.1:554/h264Preview_01_sub' -length 16.8 -print
```

## Available options

```sh
Usage of motion-speed
  -command string
        command line to run after motion (ex: echo '{{.Date}} {{.Duration}} {{.Speed}})'
  -length float
        Length of the camera view
  -motion-threshold float
        Motion threshold % (default 0.5)
  -mqtt
        Enable MQTT
  -mqtt-broker string
        MQTT Broker
  -mqtt-ca-file string
        MQTT CA file
  -mqtt-certificate-file string
        MQTT Certificate file
  -mqtt-client-id string
        MQTT Client ID (default "motion-speed")
  -mqtt-key-file string
        MQTT Key file
  -mqtt-port string
        MQTT Port
  -print
        print motion
  -print-format string
        print format
  -print-json
        print motion as json
  -save-frames
        Save start/end frames
  -video-file string
        Video file
  -video-url string
        Video url
```

## Contributing

PRs are always welcome! Before undertaking a major change, consider opening an issue for some discussion

## License

motion-speed is licensed under GNU GPLv3 License. You can find the complete text in [LICENSE](LICENSE)
