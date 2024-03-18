# A Practical Guide To Building A Push Notification Client

## Summary

This document aims to be a guide to implementing a notifications client in your language and framework of choice. The examples below are from this repositories integration tests (written for Node.js), which will need some adaptation to work in a React Native context and even more adaptation for Swift and Kotlin.

## Generating a client

The Notification Server uses Protobuf/[Connect](https://connectrpc.com/docs/introduction/) for service definitions and contracts. The service definition is published [here](https://buf.build/xmtp/example-notification-server-go). This can be used to generate clients in a range of languages. You may wish to publish your own version of the contract to be used by your client, and this will be necessary if you change any of the protobuf contracts.

To generate a Typescript service client, create a `buf.gen.yaml` file in your project root like this:

```yaml
version: v1
plugins:
  - name: es
    out: gen
    opt: target=ts
  - name: connect-web
    out: gen
    opt: target=ts
```

You can then follow the Local Generation instructions [here](https://connectrpc.com/docs/web/generating-code/#local-generation) to install the required packages that will enable you to run `buf generate buf.build/xmtp/example-notification-server-go` and generate the client code.

You can also use Buf Remote Plugins, which do not have any local dependencies other than the Buf CLI. See an example here [here](../proto/buf.gen.yaml), paying particular attention to the client code.

You can create a client instance in your code using your generated service definitions.

`client.ts`

```ts
import { createPromiseClient } from "@connectrpc/connect";
import { Notifications } from "./gen/notifications/v1/service_connect";
import { createConnectTransport } from "@connectrpc/connect-web";

export function createNotificationClient() {
  const transport = createConnectTransport({
    baseUrl: config.notificationServerUrl,
  });

  return createPromiseClient(Notifications, transport);
}
```

This will export a [Connect Client](https://connectrpc.com/docs/web/using-clients/#promises) with types matching the backend schema.

## Register your installation

This example uses Firebase for both iOS and Android push notifications. Firebase provides easy methods for getting an `installationId` and `deviceToken` for the application. If you use a different push notifications service, any opaque string that is consistent for the lifetime of an install and unique between app installations will suffice as an `installationId`. `deviceToken` can be whatever is used in your notification server's delivery service to send a notification.

```ts
import installations from "@react-native-firebase/installations";
import messaging from "@react-native-firebase/messaging";

async function register() {
  // See example above for implementation of this function
  const client = await createNotificationClient();
  // Get the FCM device token
  const deviceToken = await messaging().getToken();
  // Get the FCM installationId
  const installationId = await installations().getId();
  await client.registerInstallation(
    {
      installationId,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: deviceToken,
          case: "firebaseDeviceToken",
        },
      },
    },
    {}
  );
}
```

The client should re-register tokens periodically. A good rule of thumb might be to run the above code on app startup, so long as the device has not been registered in the past 24 hours.

## Subscribe to topics

Once your application has an instance of the `xmtp` client, you will want to subscribe to any topic to which you want to send push notifications.

This is an opinionated example that uses silent notifications for intro and invite topics on iOS and regular notifications for conversation messages.

```ts
import {
  Client,
  buildUserIntroTopic,
  buildUserInviteTopic,
} from "@xmtp/xmtp-js";
import { type PromiseClient } from "@connectrpc/connect";
import { Notifications } from "./gen/notifications/v1/service_connect";
import {
  Subscription,
  Subscription_HmacKey,
} from "./gen/notifications/v1/service_pb";

export async function subscribeToTopics(
  // The installationId we want to apply the subscription to
  installationId: string,
  // An XMTP Client. May require slight modifications when run in React Native
  xmtpClient: Client,
  // A notifications server client, like the one generated above.
  notificationClient: PromiseClient<typeof Notifications>,
  // We want to handle iOS subscriptions slightly differently because we can't filter regular notifications on the client
  isIos: boolean
) {
  // Only subscribe to notifications which have a consent state of allowed
  // to protect users from SPAM notifications
  const consentedConversations = (await xmtpClient.conversations.list()).filter(
    (c) => c.consentState === "allowed"
  );

  // Get the HMAC Keys for all conversations where the keys exist
  const hmacKeys = (
    await xmtpClient.keystore.getV2ConversationHmacKeys({
      topics: consentedConversations.map((c) => c.topic),
    })
  ).hmacKeys;

  // Convert the conversations to subscriptions
  const conversationSubscriptions = consentedConversations.map(
    (c) =>
      new Subscription({
        topic: c.topic,
        // V1 conversations don't have isSender support.
        // Use data only notifications here for iOS
        isSilent: c.conversationVersion === "v1" && isIos,
        hmacKeys: hmacKeys[c.topic]?.values.map(
          (hmacKey) =>
            new Subscription_HmacKey({
              key: hmacKey.hmacKey,
              thirtyDayPeriodsSinceEpoch: hmacKey.thirtyDayPeriodsSinceEpoch,
            })
        ),
      })
  );

  const inviteAndIntroSubscriptions: Subscription[] = [
    // Intro topic for new V1 conversations
    new Subscription({
      topic: buildUserIntroTopic(xmtpClient.address),
      isSilent: isIos,
    }),
    // Invite topic for new V2 conversations
    new Subscription({
      topic: buildUserInviteTopic(xmtpClient.address),
      isSilent: true,
    }),
  ];

  await notificationClient.subscribeWithMetadata({
    installationId,
    subscriptions: conversationSubscriptions.concat(
      inviteAndIntroSubscriptions
    ),
  });
}
```

Once the client is registered and the topics are subscribed, you should start receiving notifications from the push server.

## Revoke access on log out

If your app has some ability to log out, or switch accounts, you will want to revoke access for push notifications on that action. This can be accomplished with something like the following code:

```ts
async function revoke(installationId: string): Promise<void> {
  await subscriptionClient.deleteInstallation({
    installationId,
  });
}
```

## Listen for push notifications

Each notification has three fields in the data payload that are useful for decrypting the message.

1. `topic`
2. `encryptedMessage`
3. `messageType`

In order to decrypt a message you must find the matching conversation for the message and then call `conversation.decodeMessage`.

_TODO: Add code samples for decoding messages_

### Updating the conversation list

_TODO: Add code samples for updating conversation list_

### Types of notifications

_TODO: Add code samples for handling different notification types differently_

### How to build a high-quality client

- You probably will want to set up a per-address [notification channel](https://developer.android.com/develop/ui/views/notifications/channels&sa=D&source=docs&ust=1670358576222497&usg=AOvVaw0Iw1wSN2CR-pPhCX5tCLQF) for Android. This will make it easier for users to filter certain notification types in their app-level settings.
- [Requiring the device to be unlocked before displaying the notification](https://developer.android.com/develop/ui/views/notifications#ActionsRequireUnlockedDevice) likely makes the most sense from a privacy perspective, but that's your product decision.
- [Expandable notifications](https://developer.android.com/develop/ui/views/notifications/expanded) feel like a superior UX.
  /
