FROM alpine:edge AS build

RUN apk add --no-cache go build-base opencv-dev
WORKDIR /build

COPY go.mod go.sum /build
RUN go mod download

COPY . /build

RUN go build

FROM alpine:edge

RUN apk add --no-cache opencv libopencv_aruco libopencv_photo libopencv_video

COPY --from=build /build/motion-speed /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/motion-speed"]
CMD ["-help"]
