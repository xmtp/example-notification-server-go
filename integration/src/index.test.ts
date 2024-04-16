import { serve } from "bun";
import { expect, test, beforeEach, afterAll, describe } from "bun:test";
import { createNotificationClient, randomClient } from ".";
import { buildUserInviteTopic, fromNanoString } from "@xmtp/xmtp-js";
import type { NotificationResponse } from "./types";
import { fetcher } from "@xmtp/proto";
const { b64Decode } = fetcher;

const PORT = 7777;

describe("notifications", () => {
  let onRequest = (req: NotificationResponse) =>
    console.log("No request handler set for", req);
  // Set up a server to receive messages from the HttpDelivery service
  const server = serve({
    port: PORT,
    async fetch(req: Request) {
      const body = (await req.json()) as NotificationResponse;
      onRequest(body);
      return new Response("", { status: 200 });
    },
    // biome-ignore lint/suspicious/noExplicitAny: <explanation>
  } as any);

  afterAll(() => {
    server.stop();
  });

  const waitForNextRequest = (
    timeoutMs: number
  ): Promise<NotificationResponse> =>
    new Promise((resolve, reject) => {
      onRequest = (body) => resolve(body);
      setTimeout(reject, timeoutMs);
    });

  test("conversation invites", async () => {
    const alix = await randomClient();
    const bo = await randomClient();
    const alixNotificationClient = createNotificationClient();
    await alixNotificationClient.registerInstallation({
      installationId: alix.address,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });
    const alixInviteTopic = buildUserInviteTopic(alix.address);
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.address,
      subscriptions: [
        {
          topic: alixInviteTopic,
          isSilent: true,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(10000);
    await alix.conversations.newConversation(bo.address);
    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeString();
    expect(notification.message.content_topic).toEqual(alixInviteTopic);
    expect(notification.message.message).toBeString();
    expect(notification.subscription.is_silent).toBeTrue();
    expect(notification.installation.delivery_mechanism.token).toEqual("token");
    expect(notification.message_context.message_type).toEqual("v2-invite");
  });

  test("hmac keys", async () => {
    const alix = await randomClient();
    const bo = await randomClient();
    const alixNotificationClient = createNotificationClient();
    await alixNotificationClient.registerInstallation({
      installationId: alix.address,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });
    const conversation = await alix.conversations.newConversation(bo.address);
    const hmacKeys = await alix.keystore.getV2ConversationHmacKeys({
      topics: [conversation.topic],
    });
    const matchingKeys = hmacKeys.hmacKeys[conversation.topic].values.map(
      (v) => ({
        thirtyDayPeriodsSinceEpoch: v.thirtyDayPeriodsSinceEpoch,
        key: v.hmacKey,
      })
    );
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.address,
      subscriptions: [
        {
          topic: conversation.topic,
          isSilent: false,
          hmacKeys: matchingKeys,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(10000);
    await conversation.send("This should never be delivered");
    const boConversation = await bo.conversations.newConversation(alix.address);
    const boMessage = await boConversation.send("This should be delivered");
    expect(boConversation.topic).toEqual(conversation.topic);

    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeString();
    expect(notification.message.content_topic).toEqual(conversation.topic);
    expect(notification.message.message).toBeString();
    expect(notification.subscription.is_silent).toBeFalse();
    expect(notification.installation.delivery_mechanism.token).toEqual("token");

    const decryptedMessage = await boConversation.decodeMessage({
      timestampNs: notification.message.timestamp_ns.toString(),
      message: b64Decode(notification.message.message),
      contentTopic: notification.message.content_topic,
    });

    expect(decryptedMessage.content).toEqual("This should be delivered");
  });
});
