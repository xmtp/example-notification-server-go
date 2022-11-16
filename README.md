# example-notification-server-go

Example push notification server, written in Golang

## Project status

This repo is very much a WIP, and some important features are not yet completed.

| Feature               | Status | Notes                                                        |
| --------------------- | ------ | ------------------------------------------------------------ |
| Installations Service | 🟢     | Some minor revisions will be needed, but looking pretty good |
| Subscriptions Service | 🟢     |                                                              |
| XMTP Worker           | 🟡     | Simple implementation developed. Needs more testing          |
| API Server            | 🟡     | No tests. Maybe it works 🤷🏼‍♂️                                  |
| Delivery Service      | 🛑     | Implementation is only a stub                                |

## Prerequisites

1. Go 1.18
2. Docker and Docker Compose

## Local Setup

To start the XMTP service and database, run:

```sh
./dev/up
```

You should then be able to build the server using:

```sh
./dev/build
```

## Usage

### Running the server

The server can be run using the `./dev/run` script. Both the `worker` (which listens for new messages on the XMTP network and sends push notifications) and the `api` service (which handles HTTP/GRPC requests) are optional, but are recommended to be both enabled in local development. In a deployed environment it may be more desirable to split these services up so that you can have N instances of `api` and a single `worker`.

```sh
## Only has to be run once
./dev/up
source .env
./dev/run --worker --api
```

### Command line options

To see a full list of command line options run

```sh
./dev/run --help
```

### Generating code

If you have made a change to the files in the `proto` folder, you will need to regenerate the related Go code. You can do that with:

```sh
./dev/gen-proto
```

All required libraries should be installed as part of that process. YMMV.

### Testing the API

The API supports plain JSON and can be used via CURL

```sh
./dev/run --api
curl \
    --header "Content-Type: application/json" \
    --data '{"installationId": "123", "deliveryMechanism": {"apnsDeviceToken": "foo"}}' \
    http://localhost:8080/notifications.v1.Notifications/RegisterInstallation
```

### Running the tests

Test files must be run serially right now, due to the shared database instance which is wiped after most tests.

```sh
go test -p 1 ./...
```

## Extending the server

The implementations of the `Installations` service and the `Delivery` service are quite naive. For a production application you will likely want to replace with a more robust set of tools for managing device tokens and sending notifications idempotently. To do that, you would modify `cmd/server/main.go` and replace those service interfaces with your custom implementation.

The `Subscriptions` service has simpler requirements and will be developed to the point of suitability in a production environment.
