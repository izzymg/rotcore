# This Dockerfile builds each executable and puts them under /app
# Building the image creates all of the executables
# Running a container from image will copy them into the volume specified in docker run

# Bind mount /data to ./deploy/bin

# Go RTC app
FROM golang:1.14 AS rtc-builder

ARG GIT_TOKEN

WORKDIR /app

COPY ./rtc .

# RTC Go needs github token
RUN git config --global url."https://${GIT_TOKEN}:@github.com/".insteadOf "https://github.com/" && \
go build -o ./rtc

# Rust KBM app
FROM rust:1.40 AS kbm-builder

WORKDIR /app

COPY ./kbm .

# Rust builds into ./target/(release/debug)/(app), copy back to /app/(app)
RUN cargo build --release && cp ./target/release/kbm /app/kbm

# C Streamer app
FROM alpine:edge AS streamer-builder

WORKDIR /app
COPY ./streamer .

# Alpine needs musl-dev due to its smaller libc version lacking headers
RUN apk add --no-cache \
    musl-dev \
    gcc \
    make \
    glib-dev \
    gst-plugins-base \
    gst-plugins-good \
    gst-plugins-bad \
    gst-plugins-ugly \
    gstreamer-dev && \
    make

FROM alpine:edge

VOLUME /data

WORKDIR /app

COPY --from=rtc-builder /app/rtc .
COPY --from=kbm-builder /app/kbm .
COPY --from=streamer-builder /app/streamer .

RUN mkdir -p /data

CMD ["/bin/ash", "-c", "cp /app/* /data"]