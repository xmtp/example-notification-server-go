# Bun to Node.js Migration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate integration tests from Bun to Node.js + TypeScript using Koa (HTTP server) and Vitest (test runner).

**Architecture:** Replace all Bun-specific APIs (`bun.serve`, `bun:test`) with Node-compatible equivalents. The test structure and logic remain identical — only the runtime, server, and test framework change.

**Tech Stack:** Node.js 22, Koa, Vitest, TypeScript, npm

---

### Task 1: Update package.json

**Files:**
- Modify: `integration/package.json`

**Step 1: Rewrite package.json**

Replace the entire contents of `integration/package.json` with:

```json
{
  "name": "integration",
  "type": "module",
  "scripts": {
    "test": "vitest run"
  },
  "devDependencies": {
    "@types/koa": "^2",
    "@types/koa__bodyparser": "^5",
    "typescript": "^5.0.0",
    "vitest": "^3"
  },
  "dependencies": {
    "@bufbuild/protobuf": "^2.2.0",
    "@connectrpc/connect": "^2.0.0",
    "@connectrpc/connect-web": "^2.0.0",
    "@koa/bodyparser": "^5",
    "@xmtp/node-sdk": "5.3.0",
    "koa": "^2",
    "viem": "^2.22.2"
  }
}
```

Key changes:
- Removed `"module": "src/index.ts"` (Bun-specific field)
- Removed `@types/bun` devDependency
- Removed `peerDependencies.typescript` (moved to devDependencies)
- Added `koa`, `@koa/bodyparser`, `@types/koa`, `@types/koa__bodyparser`
- Added `vitest`
- Added `scripts.test`

**Step 2: Delete bun.lock**

```bash
rm integration/bun.lock
```

**Step 3: Install dependencies and generate package-lock.json**

```bash
cd integration && npm install
```

Expected: `package-lock.json` created, `node_modules/` updated, no errors.

**Step 4: Commit**

```bash
git add integration/package.json integration/package-lock.json
git rm integration/bun.lock
git commit -m "chore(integration): swap bun for node dependencies

Replace @types/bun with vitest, koa, and @koa/bodyparser.
Switch from bun.lock to package-lock.json."
```

---

### Task 2: Update TypeScript and Vitest config

**Files:**
- Modify: `integration/tsconfig.json`
- Create: `integration/vitest.config.ts`

**Step 1: Rewrite tsconfig.json**

Replace the entire contents of `integration/tsconfig.json` with:

```json
{
  "compilerOptions": {
    "lib": ["ESNext"],
    "target": "ESNext",
    "module": "ESNext",
    "moduleDetection": "force",
    "allowJs": true,

    "moduleResolution": "bundler",
    "verbatimModuleSyntax": false,
    "noEmit": true,

    "strict": true,
    "skipLibCheck": true,
    "noFallthroughCasesInSwitch": true,

    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noPropertyAccessFromIndexSignature": false
  }
}
```

Key changes from original:
- Removed `"jsx": "react-jsx"` (not used)
- Removed `"allowImportingTsExtensions": true` (we'll remove `.ts` extensions from imports)
- Kept `"moduleResolution": "bundler"` — Vitest uses Vite which is a bundler, so this is the correct setting

**Step 2: Create vitest.config.ts**

Create `integration/vitest.config.ts`:

```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    testTimeout: 30_000,
    hookTimeout: 30_000,
  },
});
```

**Step 3: Commit**

```bash
git add integration/tsconfig.json integration/vitest.config.ts
git commit -m "chore(integration): update tsconfig and add vitest config

Remove Bun-specific tsconfig options. Add vitest config with
30s timeouts for XMTP network operations."
```

---

### Task 3: Migrate test file from Bun to Vitest + Koa

**Files:**
- Modify: `integration/src/index.test.ts`

This is the core migration task. The test logic stays identical — only the imports, HTTP server, and assertion matchers change.

**Step 1: Rewrite index.test.ts**

Replace the entire contents of `integration/src/index.test.ts` with:

```ts
import Koa from "koa";
import { bodyParser } from "@koa/bodyparser";
import { expect, test, afterAll, describe } from "vitest";
import { createNotificationClient, randomClient } from ".";
import type { NotificationResponse } from "./types";

const PORT = 7777;

describe("notifications", () => {
  let onRequest = (req: NotificationResponse) =>
    console.log("No request handler set for", req);

  // Set up a Koa server to receive messages from the HttpDelivery service
  const app = new Koa();
  app.use(bodyParser());
  app.use(async (ctx) => {
    onRequest(ctx.request.body as NotificationResponse);
    ctx.status = 200;
  });
  const server = app.listen(PORT);

  afterAll(() => {
    server.close();
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
      installationId: alix.installationId,
      deliveryMechanism: {
        deliveryMechanismType: {
          value: "token",
          case: "apnsDeviceToken",
        },
      },
    });

    const alixInviteTopic = `/xmtp/mls/1/g-${alix.installationId}/proto`;
    await alixNotificationClient.subscribeWithMetadata({
      installationId: alix.installationId,
      subscriptions: [
        {
          topic: alixInviteTopic,
          isSilent: true,
        },
      ],
    });

    const notificationPromise = waitForNextRequest(1000);
    await alix.conversations.createDm(bo.inboxId);
    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeTypeOf("string");
    expect(notification.message.content_topic).toEqual(alixInviteTopic);
    expect(notification.message.message).toBeTypeOf("string");
    expect(notification.subscription.is_silent).toBe(true);
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

    const boGroup = await bo.conversations.createGroup([alix.inboxId]);

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
    await alixGroup.sendText("This should never be delivered");
    await boGroup.sendText("This should be delivered");

    const notification = await notificationPromise;

    expect(notification.idempotency_key).toBeTypeOf("string");
    expect(notification.message.content_topic).toEqual(topic);
    expect(notification.message.message).toBeTypeOf("string");
    expect(notification.subscription.is_silent).toBe(false);
    expect(notification.installation.delivery_mechanism.token).toEqual("token");
  });
});
```

Changes from original:
- Lines 1-2: `import { serve } from "bun"` → Koa + bodyParser imports
- Line 3: `import { ... } from "bun:test"` → `import { ... } from "vitest"`
- Lines 12-20: `bun.serve({...})` → Koa app with bodyParser middleware
- Line 23: `server.stop()` → `server.close()`
- Line 63: `toBeString()` → `toBeTypeOf("string")`
- Line 65: `toBeString()` → `toBeTypeOf("string")`
- Line 66: `toBeTrue()` → `toBe(true)`
- Line 120: `toBeString()` → `toBeTypeOf("string")`
- Line 122: `toBeString()` → `toBeTypeOf("string")`
- Line 123: `toBeFalse()` → `toBe(false)`

**Step 2: Commit**

```bash
git add integration/src/index.test.ts
git commit -m "feat(integration): migrate tests from bun:test to vitest + koa

Replace bun.serve with Koa HTTP server.
Replace bun:test imports with vitest.
Replace Bun-specific matchers with vitest equivalents."
```

---

### Task 4: Fix source file imports

**Files:**
- Modify: `integration/src/index.ts`

**Step 1: Remove .ts extension from import**

In `integration/src/index.ts`, change line 7:

```ts
// Before:
import {
  Notifications,
  SubscriptionSchema,
  Subscription_HmacKeySchema,
} from "./gen/notifications/v1/service_pb.ts";

// After:
import {
  Notifications,
  SubscriptionSchema,
  Subscription_HmacKeySchema,
} from "./gen/notifications/v1/service_pb";
```

This is the only source file change needed. All other files (`config.ts`, `types.ts`, `gen/`) require no modifications.

**Step 2: Commit**

```bash
git add integration/src/index.ts
git commit -m "fix(integration): remove .ts extension from import path

Not needed without Bun's allowImportingTsExtensions."
```

---

### Task 5: Update Dockerfile

**Files:**
- Modify: `integration/Dockerfile`

**Step 1: Rewrite Dockerfile**

Replace the entire contents of `integration/Dockerfile` with:

```dockerfile
FROM node:22-slim AS base
WORKDIR /usr/app

# Install dependencies into temp directory
# This will cache them and speed up future builds
FROM base AS install
COPY package.json package-lock.json ./
RUN npm ci

# The real release artifact
FROM base
COPY --from=install /usr/app/node_modules node_modules
COPY . .

EXPOSE 7777/tcp
CMD ["npx", "vitest", "run"]
```

Key changes:
- Base image: `oven/bun:1` → `node:22-slim`
- Lockfile: `bun.lock` → `package-lock.json`
- Install: `bun install --frozen-lockfile` → `npm ci`
- Removed `USER bun` (not applicable)
- Changed `ENTRYPOINT` → `CMD` (allows compose to override cleanly)

**Step 2: Commit**

```bash
git add integration/Dockerfile
git commit -m "chore(integration): update Dockerfile from bun to node

Use node:22-slim base image, npm ci, and vitest entrypoint."
```

---

### Task 6: Update Docker Compose

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Update integration service command**

In `docker-compose.yml`, change the integration service command (lines 66-68):

```yaml
# Before:
    command:
      - bun
      - test

# After:
    command:
      - npx
      - vitest
      - run
```

No other changes to `docker-compose.yml`.

**Step 2: Commit**

```bash
git add docker-compose.yml
git commit -m "chore: update docker-compose integration command to vitest"
```

---

### Task 7: Run integration tests end-to-end

**Step 1: Ensure Docker services are running**

```bash
./dev/up
```

**Step 2: Run integration tests**

```bash
./dev/integration
```

Expected: Both tests pass ("conversation invites" and "hmac keys").

**Step 3: If tests fail, debug and fix**

Common issues to check:
- Koa bodyParser not parsing JSON correctly → verify `Content-Type` header from Go server
- Import resolution errors → check that `.ts` extension was removed from `index.ts`
- Timeout errors → increase `testTimeout` in `vitest.config.ts`
- `@xmtp/node-sdk` native bindings → may need to ensure the Docker image has required system libs (node:22-slim should be sufficient)

**Step 4: Final commit (if any fixes were needed)**

```bash
git add -A
git commit -m "fix(integration): address issues found during integration testing"
```
