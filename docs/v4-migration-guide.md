# Decentralized Network Migration Guide

## Who is this for

This guide is for developers already running a notification server (either this code directly or a fork) in a deployed environment. This guide walks you through how to handle the change without interrupting deliveries for clients.

## What's changing

On [CUTOVER DATE] all XMTP clients will stop writing to the current XMTP `v3` network and begin writing to the decentralized `v4` network. The current version of the notification server is configured to read from either network. Before the cutover date, you will need to configure your notification server to read from a `v4` endpoint or message deliveries will stop.

## Preparing for the migration

The latest version of the notification server is designed to smooth over the transition by translating messages from `v4` formats to `v3` formats. This means that **no client changes are required to upgrade your notification server to the latest version**. Even if you are connected to the `v4` network, legacy clients can receive messages in the format they already expect.

## Required changes before the cutover


### 1. Update your client

[TODO: Describe changes to client methods for receiving v4 payload format notifications]

### 2. Switch your development environment to `testnet`

If you currently use the `dev` environment for local development and test versions of your application, you should configure those clients to use the `testnet` network instead.

Once you have rolled out that change, configure the notification server you use for your development environment with the following flags

- Set the `--listener-type` command line flag or `LISTENER_TYPE` environment variable to `v4`. This tells the notification server to expect payloads from the XMTP network to arrive in a `v4` format
- Set the `--xmtp-address` command line flag or `XMTP_GRPC_ADDRESS` environment variable to `https://grpc.testnet.xmtp.network:443`. This tells the notification server to connect to the new testnet nodes.

Testnet is a wholly new environment and old messages do not carry over. Any clients previously on the `dev` network will need to sign in again and start with a clean messaging history.

### 3. Switch your production environment to `mainnet`

In mainnet, XMTP is running a continuous migration of new messages from `v3` to `v4` up until the cutover. That means that your notification server can start receiving and forwarding messages from the `v4` network today, before clients start talking directly to the `v4` network on [CUTOVER DATE].

For the notification server you use in your production environment, make the following configuration changes:

- Set the `--listener-type` command line flag or `LISTENER_TYPE` environment variable to `v4`. This tells the notification server to expect payloads from the XMTP network to arrive in a `v4` format.
- Set the `--xmtp-address` command line flag or `XMTP_GRPC_ADDRESS` environment variable to `https://grpc.mainnet.xmtp.network:443`. This tells the notification server to connect to mainnet.

## Performance and latency

The `v4` network has comparable latency and throughput to `v3`. Until [CUTOVER DATE] messages sent to `v3` production must travel from `production` -> migration service -> `mainnet`. That additional hop adds approximately 2 seconds of delay before a message is received by the push server. After [CUTOVER DATE] latency will return to normal.

## Breaking changes

- Previous versions of the example notification server did not validate whether subscribed topics matched the expected XMTP topic format. You could subscribe to `/foo` previously and it would be accepted. Version 2.0 and above will assert that all subscribed topics must match the expected topic format and error on subscriptions that contain invalid topics. This is mostly relevant if you are writing integration tests against your notification server.

## Under the hood changes

As XMTP moves from the `v3` network to the `v4` decentralized backend there are a few important changes to the protocol that impact the internals of the notification server.

- New binary format for topics (convertible in both directions). A database migration upgrades all previously saved topics.
- New wrapper envelope types (`v4` envelopes can be converted back to `v3` envelopes)
- New URL to conenct to the `mainnet` (f.k.a `production`) and `testnet` (f.k.a `dev`) environments
