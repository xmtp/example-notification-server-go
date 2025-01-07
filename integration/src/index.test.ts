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

    const alixInviteTopic = `/xmtp/mls/1/g-${alix.accountAddress}/proto`;
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.accountAddress,
      subscriptions: [
        {
          topic: alixInviteTopic,
          isSilent: true,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(1000);
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
      installationId: alix.installationId,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });

    const boGroup = await bo.conversations.newGroup([alix.accountAddress]);

    expect((await alix.conversations.list()).length).toEqual(0);
    await alix.conversations.syncAll();
    const alixGroups = await alix.conversations.list();
    expect(alixGroups.length).toEqual(1);
    const alixGroup = alixGroups[0];

    const hmacKeys = alix.conversations.hmacKeys();
    expect(Object.keys(hmacKeys).length).toEqual(1);
    const conversationHmacKeys = hmacKeys[alixGroup.id];
    expect(conversationHmacKeys.length).toEqual(3);

    const matchingKeys = conversationHmacKeys.map((v) => ({
      thirtyDayPeriodsSinceEpoch: Number(v.epoch),
      key: Uint8Array.from(v.key),
    }));
    const topic = `/xmtp/mls/1/g-${alixGroup.id}/proto`;
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.installationId,
      subscriptions: [
        {
          topic,
          isSilent: false,
          hmacKeys: matchingKeys,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(10000);
    await alixGroup.send("This should never be delivered");
    await boGroup.send("This should be delivered");

    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeString();
    expect(notification.message.content_topic).toEqual(topic);
    expect(notification.message.message).toBeString();
    expect(notification.subscription.is_silent).toBeFalse();
    expect(notification.installation.delivery_mechanism.token).toEqual("token");
  });
});
