# A Practical Guide To Building A Push Notification Client

## Summary

This document aims to be a guide to implementing a notifications client in your language and framework of choice. I have implemented a [working example in React Native](https://github.com/xmtp/example-chat-react-native/pull/6), and all code samples will be in a RN context. With a bit of creativity, however, you should be able to figure out a way to implement a notifications client in your language of choice.

## Generating a client

The Notification Server uses Protobuf/[Connect](https://connect.build/docs/introduction) for service definitions and contracts. The service definition is published [here](https://buf.build/nickxmtp/example-notification-server/docs/main:notifications.v1). This can be used to generate clients in a range of languages. You may wish to publish your own version of the contract to be used by your client, and this will be necessary if you change any of the protobuf contracts.

To generate a Typescript service client, create a `buf.gen.yaml` file in your project root like this:

```yaml
version: v1
plugins:
  - name: es
    out: gen
    # With target=ts, we generate TypeScript files.
    # Use target=js+dts to generate JavaScript and TypeScript declaration files
    # like remote generation does.
    opt: target=ts
  - name: connect-web
    out: gen
    # With target=ts, we generate TypeScript files.
    opt: target=ts
```

You can then follow the Local Generation instructions [here](https://connect.build/docs/web/generating-code#local-generation) to install the required packages that will enable you to run `buf generate buf.build/nickxmtp/example-notification-server` and generate the client code.

You can create a client instance in your code with something like this:

`client.ts`

```ts
import { Notifications } from "../gen/service_connectweb";
import {
  createConnectTransport,
  createPromiseClient,
} from "@bufbuild/connect-web";

const transport = createConnectTransport({
  baseUrl: process.env.API_URL,
});

// Here we make the client itself, combining the service
// definition with the transport.
const client = createPromiseClient(Notifications, transport);

export default client;
```

This will export a [Connect Client](https://connect.build/docs/web/using-clients#promises) with types matching the backend schema.

## Register your installation

In my example, I am using Firebase for both iOS and Android push notifications. Firebase provides easy methods for getting an `installationId` and `deviceToken` for the application. If you use a different push notifications service, any opaque string that is consistent for the lifetime of an install and unique between app installations will suffice as an `installationId`. `deviceToken` can be whatever is used in your notification server's delivery service to send a notification (for example, an APNS token or OneSignal device token).

I have a [React Hook](https://github.com/xmtp/example-chat-react-native/blob/nm/add-firebase/hooks/useRegister.ts#L26) that does something like the following code on application startup:

```ts
import installations from "@react-native-firebase/installations";
import messaging from "@react-native-firebase/messaging";

async function register() {
  const deviceToken = await messaging().getToken();
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

The Notification Server will expire device tokens that have not been updated in a configurable period of time, so the client should re-register tokens periodically. A good rule of thumb might be to run the above code on app startup, so long as the device has not been registered in the past 24 hours.

## Subscribe to topics

Once your application has an instance of the `xmtp` client, you will want to subscribe to all topics so that the server will know what messages to forward to the client.

```ts
import { Client } from "@xmtp/xmtp-js";
import installations from "@react-native-firebase/installations";
import {
  buildUserIntroTopic,
  buildUserInviteTopic,
  //@ts-ignore
} from "@xmtp/xmtp-js/dist/cjs/src/utils";

export const updateSubscriptions = async (xmtp: Client) => {
  const conversations = await xmtp.conversations.list();
  const installationId = await installations().getId();
  const convoTopics = conversations.map((convo) => convo.topic);
  const topics = [
    ...convoTopics,
    buildUserIntroTopic(xmtp.address), // Used to receive V1 introductions
    buildUserInviteTopic(xmtp.address), // Used to receive V2 invites
  ];

  await apiClient.subscribe(
    {
      installationId,
      topics,
    },
    {}
  );
};
```

Once the client is registered and the topics are subscribed, you should start receiving messages from the push server.

## Revoke access on log out

If your app has some ability to log out, or change accounts, you will want to revoke access for push notifications on that action. This can be accomplished with something like the following code:

```ts
async function revoke(installationId: string): Promise<void> {
  await subscriptionClient.deleteInstallation({
    installationId,
  });
}
```

## Listen for push notifications

Each notification has two fields in the data payload that are required to decrypt the message.

1. `topic`
2. `encryptedMessage`

In order to decrypt a message you must find the matching conversation for the message and then call `conversation.decodeMessage`.

```ts
for (const conversation of await client.conversations.list()) {
  if (conversation.topic === topic) {
    return conversation.decodeMessage({
      contentTopic: topic,
      message: message as unknown as any, // There is some weirdness with the generated types here
    });
  }
}
```

Once decoded, you can create a local notification using the framework of your choice to display it to the user. You may also choose to enrich the payload by looking up the ENS reverse record for the sender address, and maybe adding an ENS avatar or Blockie as a profile image.

In my example, I am using the React Native Firebase SDK to listen for notifications and handle them. I've implemented some basic caching of the conversation list and XMTP client to limit the number of network requests required to decrypt the message.

You can see an extremely hacky but functional implementation [here](https://github.com/xmtp/example-chat-react-native/blob/nm/add-firebase/lib/notifications.ts).

### Updating the conversation list

Not all subscribed messages are meant to be displayed to the user. Topics with the prefix `/xmtp/0/invite-` are invitations, and should be used as an opportunity to refresh the conversation list for later. Topics with the prefix `/xmtp/0/intro-` are introduction messages and can be displayed to the user, but also may indicate the beginning of a new conversation. Those should trigger a refresh of the conversation list.

### Types of notifications

The Notification Server, as currently configured, sends both iOS and Android notifications as "background notifications". That is to say that it has no Notification payload, and has the `content-available` flag set to true on iOS and the `priority` set to 5 on Android.

This is probably the right approach for Android, but on iOS runs the risk of getting rate-limited when operating at scale. To implement this as regular notifications that do not get rate limited as aggressively, you could include a Notification element in the payload and set the `mutable-content` flag to true. The challenge here is that you will then require a [Notification Service Extension](https://developer.apple.com/documentation/usernotifications/modifying_content_in_newly_delivered_notifications) to handle the decryption of the content. You can read more about [Notification Service Extensions](https://www.strv.com/blog/app-extensions-introduction-to-notification-service-engineering) here.

### How to build a high quality client

- Some additional data fields will likely be useful in a production service. `id`, `type`, and `version` could all help build a high quality client. Those can be configured in the Delivery Service. For example, you may want to dedupe notifications by ID, route to different handlers using type, and use the version field to ensure compatibility between the client and server notification schema.
- `client.conversations.list()` can be a slow and expensive call with lots of heavy cryptographic operations. Especially for users with many ongoing chats. Caching the conversation list and only refreshing when necessary would make the notification handler far more performant.
- You probably will want to setup a per-address [notification channel](https://developer.android.com/develop/ui/views/notifications/channels&sa=D&source=docs&ust=1670358576222497&usg=AOvVaw0Iw1wSN2CR-pPhCX5tCLQF) for Android. This will make it easier for users to filter certain notification types in their app level settings.
- [Requiring the device to be unlocked before displaying the notification](https://developer.android.com/develop/ui/views/notifications#ActionsRequireUnlockedDevice) likely makes the most sense from a privacy perspective, but that's your product decision.
- [Expandable notifications](https://developer.android.com/develop/ui/views/notifications/expanded) feel like a superior UX
