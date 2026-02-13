import { Client, type Signer, IdentifierKind } from "@xmtp/node-sdk";
import { createWalletClient, http, toBytes } from "viem";
import { mainnet } from "viem/chains";
import { privateKeyToAccount, generatePrivateKey } from "viem/accounts";
import { createClient, type Client as ConnectClient } from "@connectrpc/connect";
import {
  Notifications,
  SubscriptionSchema,
  Subscription_HmacKeySchema,
} from "./gen/notifications/v1/service_pb.ts";
import { createConnectTransport } from "@connectrpc/connect-web";
import { config } from "./config";
import { create } from "@bufbuild/protobuf";
import { getRandomValues } from "node:crypto";

export function randomWallet() {
  const account = privateKeyToAccount(generatePrivateKey());
  return createWalletClient({
    account,
    chain: mainnet,
    transport: http(),
  });
}

export async function randomClient() {
  const wallet = randomWallet();
  const signer: Signer = {
    type: "EOA",
    getIdentifier: () => ({
      identifier: wallet.account.address,
      identifierKind: IdentifierKind.Ethereum,
    }),
    signMessage: async (message) => {
      const signature = await wallet.signMessage({ message });
      return toBytes(signature);
    },
  };

  const encKey = getRandomValues(new Uint8Array(32));
  return await Client.create(signer, {
    env: "local",
    apiUrl: config.nodeUrl,
    dbEncryptionKey: encKey,
    dbPath: `/tmp/test-${wallet.account.address}.db3`,
  });
}

export function createNotificationClient() {
  const transport = createConnectTransport({
    baseUrl: config.notificationServerUrl,
  });

  return createClient(Notifications, transport);
}

export async function subscribeToTopics(
  // The installationId we want to apply the subscription to
  installationId: string,
  // An XMTP Client. May require slight modifications when run in React Native
  xmtpClient: Client,
  // A notifications server client, like the one generated above.
  notificationClient: ConnectClient<typeof Notifications>,
  // We want to handle iOS subscriptions slightly differently because we can't filter regular notifications on the client
  isIos: boolean
) {
  // Only subscribe to notifications which have a consent state of allowed
  // to protect users from SPAM notifications
  const consentedConversations = (await xmtpClient.conversations.list()).filter(
    (c) => c.consentState() === 1
  );

  // Get the HMAC Keys for all conversations where the keys exist
  const hmacKeys = xmtpClient.conversations.hmacKeys();

  // Convert the conversations to subscriptions
  const conversationSubscriptions = consentedConversations.map((c) =>
    create(SubscriptionSchema, {
      topic: c.id,
      // V1 conversations don't have isSender support.
      // Use data only notifications here for iOS
      isSilent: false,
      hmacKeys: hmacKeys[c.id]?.map((v) =>
        create(Subscription_HmacKeySchema, {
          thirtyDayPeriodsSinceEpoch: Number(v.epoch),
          key: Uint8Array.from(v.key),
        })
      ),
    })
  );

  const inviteAndIntroSubscriptions: typeof conversationSubscriptions = [];

  await notificationClient.subscribeWithMetadata({
    installationId,
    subscriptions: conversationSubscriptions,
  });
}
