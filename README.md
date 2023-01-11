# example-notification-server-go

Example push notification server, written in Golang

## Project status

![Status](https://camo.githubusercontent.com/47c9762c88d56b96ffa436e2af994dab07f6f61f2a0388cd08be7d42b1b8fef5/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f50726f6a6563745f5374617475732d446576656c6f7065725f507265766965772d79656c6c6f77)

This project is in Developer Preview status and ready to serve as a reference for you to start building.

However, we do NOT recommend using Developer Preview software in production apps. Software in this status may change based on feedback.

Many applications will have different needs for push notifications (different delivery providers, different metadata attached to payloads, etc), and this repo is designed to be forked and customized for each application's needs.

| Feature               | Status | Notes                                                                                                                                           |
| --------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Installations Service | 🟢     | Some minor revisions will be needed, but looking pretty good                                                                                    |
| Subscriptions Service | 🟢     |                                                                                                                                                 |
| API Server            | 🟢     | Working as expected                                                                                                                             |
| XMTP Worker           | 🟢     | Needs more testing, but works                                                                                                                   |
| Delivery Service      | 🟢     | Basic implementation in place. You may want to adjust the notification payload, or add a new delivery service, to suit your application's needs |

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
./dev/run --xmtp-listener --api
```

### Running the listener and just logging the message

If you want to just log the message received from the listener and nothing more, use this branch and do the following:

````sh
./dev/up
source .env
./dev/run --xmtp-listener
```

### Command line options

To see a full list of command line options run

```sh
./dev/run --help
````

Here is the output as of 21/12/2021:

```sh
Usage:
  main [OPTIONS]

Application Options:
  -d, --db-connection-string=        Address to database [$DB_CONNECTION_STRING]
      --log-encoding=[console|json]  Log encoding (default: console) [$LOG_ENCODING]
      --log-level=[debug|info|error] log-level (default: info) [$LOG_LEVEL]
      --create-migration=            create a migration with the given name

API Options:
      --api                          Enable the GRPC API server
  -p, --api-port=                    Port for the Connect GRPC API (default: 8080) [$API_PORT]

Worker Options:
      --xmtp-listener                Enable the XMTP listener to actually send notifications. Requires APNSOptions to
                                     be configured
      --xmtp-listener-tls            Whether to connect to XMTP network using TLS
  -x, --xmtp-address=                Address (including port) of XMTP GRPC server [$XMTP_GRPC_ADDRESS]
      --num-workers=                 Number of workers used to process messages (default: 50)

APNS Options:
      --apns-enabled                 Enable APNS [$APNS_ENABLED]
      --apns-p8-certificate=         .p8 certificate for APNS [$APNS_P8_CERTIFICATE]
      --apns-key-id=                 Key ID associated with APNS credentials [$APNS_KEY_ID]
      --apns-team-id=                APNS Team ID [$APNS_TEAM_ID]
      --apns-topic=                  Topic to be used on all messages [$APNS_TOPIC]

FCM Options:
      --fcm-enabled                  Enable FCM sending [$FCM_ENABLED]
      --fcm-credentials-json=        FCM Credentials [$FCM_CREDENTIALS_JSON]
      --fcm-project-id=              FCM Project ID [$FCM_PROJECT_ID]

Help Options:
  -h, --help                         Show this help message
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

The implementations of the `Installations` service and the `Delivery` service are designed to be easily replaced. For a production application, you will likely want to replace them with a more robust set of tools for managing device tokens and sending notifications idempotently. To do that, you would modify `cmd/server/main.go` and replace those service interfaces with your custom implementation.

If you are using Firebase for push delivery, the only modifications needed (if any) may be to customize the payload sent to clients.

The `Subscriptions` service has simpler requirements and will be developed to the point of suitability in a production environment.

## Deployment

You will need to deploy your own instance of the Notification Server, with the appropriate credentials to send push notifications on behalf of your app. The deployed service will require access to a Postgres database as well as the ability to connect to the public internet.

You may choose to run both the API and Listener in a single service or as two separate services, depending on the expected load to the API server. For high traffic applications it is recommended to run the API server and Listener as separate services.

## Implementing a client

Once you have the server deployed, you will need to connect to it from your client application to register devices and subscriptions. There is a guide to help guide you in this process [here](./docs/notifications-client-guide.md).
