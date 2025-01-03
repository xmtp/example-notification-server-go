import { serve } from "bun";
import { expect, test, afterAll, describe } from "bun:test";
import { createNotificationClient, randomClient } from ".";
import type { NotificationResponse } from "./types";

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
      installationId: alix.accountAddress,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });

    const alixInviteTopic = `invite-${alix.accountAddress}`;
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.accountAddress,
      subscriptions: [
        {
          topic: alixInviteTopic,
          isSilent: true,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(10000);
    await alix.conversations.newDm(bo.accountAddress);
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
      installationId: alix.accountAddress,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });

    const alixConversation = await alix.conversations.newDm(bo.accountAddress);
    const hmacKeys = alix.conversations.hmacKeys();
    const conversationHmacKeys = hmacKeys[alixConversation.id];

    const matchingKeys = conversationHmacKeys.map((v) => ({
      thirtyDayPeriodsSinceEpoch: Number(v.epoch),
      key: Uint8Array.from(v.key),
    }));
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.accountAddress,
      subscriptions: [
        {
          topic: alixConversation.id,
          isSilent: false,
          hmacKeys: matchingKeys,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(10000);
    await alixConversation.send("This should never be delivered");
    const boConversation = await bo.conversations.newDm(alix.accountAddress);
    const boMessage = await boConversation.send("This should be delivered");
    // expect(boConversation.id).toEqual(conversation.id);

    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeString();
    expect(notification.message.content_topic).toEqual(alixConversation.id);
    expect(notification.message.message).toBeString();
    expect(notification.subscription.is_silent).toBeFalse();
    expect(notification.installation.delivery_mechanism.token).toEqual("token");

    // const decryptedMessage = await boConversation.decodeMessage({
    // timestampNs: notification.message.timestamp_ns.toString(),
    // message: b64Decode(notification.message.message),
    // contentTopic: notification.message.content_topic,
    // });

    // expect(decryptedMessage.content).toEqual("This should be delivered");
  });
});
