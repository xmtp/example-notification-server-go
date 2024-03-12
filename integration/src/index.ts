import {
  Client,
  buildUserIntroTopic,
  buildUserInviteTopic,
} from "@xmtp/xmtp-js";
import { createWalletClient, http } from "viem";
import { mainnet } from "viem/chains";
import { privateKeyToAccount, generatePrivateKey } from "viem/accounts";
import { createPromiseClient, type PromiseClient } from "@connectrpc/connect";
import { Notifications } from "./gen/notifications/v1/service_connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { config } from "./config";
import {
  Subscription,
  Subscription_HmacKey,
} from "./gen/notifications/v1/service_pb";

export function randomWallet() {
  const account = privateKeyToAccount(generatePrivateKey());
  return createWalletClient({
    account,
    chain: mainnet,
    transport: http(),
  });
}

export function randomClient() {
  const wallet = randomWallet();
  return Client.create(wallet, { env: "local", apiUrl: config.nodeUrl });
}

export function createNotificationClient() {
  const transport = createConnectTransport({
    baseUrl: config.notificationServerUrl,
  });

  return createPromiseClient(Notifications, transport);
}

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
    (c): Subscription =>
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
