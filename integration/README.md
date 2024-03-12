# integration

This package is designed to run as an integration test suite for the notification server. It may also serve as a useful reference for how to interact with the server.

These tests rely on the `HttpDelivery` Delivery Service on the node to send notifications directly back to the server to verify the notification content. Normally notifications would be sent by APNS or FCM.

It is meant to be run inside Docker.

## Usage

In the root of the repo

```bash
./dev/integration
```

## Development setup

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run src/index.ts
```
