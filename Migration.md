# Migration

This document describes how to run the Notification Server against different versions of XMTP nodes, especially moving from V3 to V4 version of the XMTP node.

## Legacy - V3

The V3 version of the XMTP node is the version of the XMTP node running in production as of 11. March 2026.
The source code for the node can be found at https://github.com/xmtp/xmtp-node-go.

This is the default node that Notification Server will expect to find on the other side.
Notification Server will try to communicate with it using the well established, legacy, V3 GRPC API, found [here](https://github.com/xmtp/proto/blob/main/proto/message_api/v1/message_api.proto).

The Notification Server will issue a `SubscribeAll` call to the V3 server and receive all messages flowing through the network.

## Decentralization (D14N) API - V4

The V4 version is the new, decentralized version of the XMTP node, whose implementation can be found at https://github.com/xmtp/xmtpd.
Soon it is a goal of XMTP network to completely switch to this node implementation for powering the network, so at some point message traffic will stop flowing through the legacy, v3 node, and will happen only on V4.

The API for the V4 node can be found [here](https://github.com/xmtp/proto/blob/main/proto/xmtpv4/message_api/message_api.proto).

### Running the Notification Server Against D14N Node

In order to run the Notification Server against the new node version, it is sufficient to specify `--d14n` CLI flag.
All remaining CLI flags are backwards compatible.

This will instruct the Notification Server to connect to the specified node, but will attempt to communicate with it using the new API, and will understand the different data types that the Node will send as a response.

It is possible to run two instances of the Notification Server, each pointed at a different network - V3 and V4.
Nodes can share the same underlying database from which they can pull installation and subscription data.

### Data Format Difference

As the two networks operate somewhat differently, they have different data formats.
Note that HTTP delivery mechanism provides more detailed contextual information about the message than APNs and FCM delivery methods.

#### APNs and FCM

Payloads sent by the Notification Server for APNs and FCM are exactly the same, running against the V3 or V4 API.

#### HTTP Delivery

If HTTP Delivery method is selected, Notification Server will POST the message payload to the specified address.

The payload is unchanged if the Notification Server is running against the V3 node.
If Notification Server is subscribed to a V4 node, the payload is slightly different, and can be found in the `message_v4` field.

If the Notification Server payload is sent to a single server, data origin can be determined based on the data receivedČ
- `message` - legacy, V3 message
- `message_v4` - D14N, V4 message

Seeing which field is populated can also make it easy on the data consumer to determine how to interpret the data.

These types are defined and can be inspected in more detail [here](pkg/interfaces/interfaces.go) - check the `SendRequest` type definition.
