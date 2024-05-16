# Overview

A multi-chatroom application implemented in Go that handles websocket connections and the transmission of messages to different rooms over a swappable message broker.

## Features

This chat application is deployable to multiple servers behind a load balancer. Each instance handles websocket connections and manages clients and rooms while broadcasting messages through a message broker like Redis, RabbitMQ, or one of your own implementation.

- Anonymous users are allowed by default but can be configured in the `.env` file with `ALLOW_ANON="false"`.
- Authentication middleware is implemented with JSON web tokens. Set `JWT_SECRET` in `.env` before use.
- This application does **not** implement a signup or login process but can verify and parse existing JSON web tokens stored in either a cookie (set `COOKIE_NAME` in `.env`) or transmitted through the request header in the `Authorization` field. The only field necessary in the token is `username`. Refer to `pkg/auth/jwt.go` for more details.

## Prerequisites

- Go 1.22+ since the project uses its routing features but it should be fairly straightforward to implement your own router if using earlier versions of Go.

- A running instance of Redis or RabbitMQ. If you do not already have a running instance, you can run `docker run -d --name redis -p 6379:6379 redis` to run a local Redis instance in Docker.

## Installation 

1. Download or clone the project to your workspace: `git clone https://github.com/davidado/go-chat-rooms`

2. Edit the Makefile and set the MAIN_PACKAGE_PATH to either `./cmd/redis` or `./cmd/rabbitmq` (default is `./cmd/redis`).

## Usage

1. Execute `make run` in the root path (same directory as the Makefile).

2. Go to http://localhost:8080/<room> where <room> can be anything you choose. 

3. Test the room by going to the same URL in another tab or browser and type away.

## Contributing

Contributions are welcome.
