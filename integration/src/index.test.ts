import { serve } from "bun";
import { expect, test, beforeEach, afterAll, describe } from "bun:test";
import { createNotificationClient, randomClient } from ".";
import { buildUserInviteTopic } from "@xmtp/xmtp-js";
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
      installationId: alix.address,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });
    console.log("Installation registered");
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
});
