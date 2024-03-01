import { Client } from "@xmtp/xmtp-js";
import { createWalletClient, http } from "viem";
import { mainnet } from "viem/chains";
import { privateKeyToAccount, generatePrivateKey } from "viem/accounts";
import { createPromiseClient } from "@connectrpc/connect";
import { Notifications } from "./gen/notifications/v1/service_connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { config } from "./config";

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
