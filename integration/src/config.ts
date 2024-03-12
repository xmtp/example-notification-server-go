function assertEnvVar(key: string): string {
  const value = process.env[key];
  if (!value) {
    throw new Error(`Missing environment variable ${key}`);
  }

  return value;
}

export const config = {
  nodeUrl: assertEnvVar("XMTP_NODE_URL"),
  notificationServerUrl: assertEnvVar("NOTIFICATION_SERVER_URL"),
} as const;
