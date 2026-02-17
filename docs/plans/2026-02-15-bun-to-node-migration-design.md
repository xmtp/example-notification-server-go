# Migrate Integration Tests from Bun to Node.js + TypeScript

## Context

The integration tests in `integration/` currently use Bun as both the runtime and test runner. We want to migrate to plain Node.js with TypeScript to reduce the dependency on Bun-specific APIs and tooling.

## Decisions

- **HTTP server:** Koa + @koa/bodyparser (replaces `bun.serve`)
- **Test runner:** Vitest (replaces `bun:test`)
- **TypeScript execution:** tsx (via Vitest)
- **Package manager:** npm (replaces bun)

## Scope of Changes

### Package & Dependencies

Remove: `@types/bun`, `bun.lock`

Add:
- `koa`, `@koa/bodyparser`, `@types/koa`, `@types/koa__bodyparser` — HTTP server
- `vitest` — test runner
- `tsx` — TypeScript execution

Keep unchanged: `@bufbuild/protobuf`, `@connectrpc/connect`, `@connectrpc/connect-web`, `@xmtp/node-sdk`, `viem`

### Test File (`index.test.ts`)

- Replace `import { serve } from "bun"` with Koa server setup
- Replace `import { ... } from "bun:test"` with `import { ... } from "vitest"`
- Replace Bun-specific matchers: `toBeString()` -> `toBeTypeOf("string")`, `toBeTrue()` -> `toBe(true)`, `toBeFalse()` -> `toBe(false)`
- Replace `server.stop()` with `server.close()`

### TypeScript Config (`tsconfig.json`)

- `moduleResolution`: `"bundler"` -> `"NodeNext"`
- `module`: `"ESNext"` -> `"NodeNext"`
- Remove `allowImportingTsExtensions`
- Add Vitest types reference

### Vitest Config (new `vitest.config.ts`)

- Set `testTimeout` high enough for XMTP operations (~30s)
- Set `hookTimeout` similarly

### Dockerfile

- Base image: `oven/bun:1` -> `node:22-slim`
- Package manager: `bun install` -> `npm install`
- Lockfile: `bun.lock` -> `package-lock.json`
- Entrypoint: `bun test` -> `npx vitest run`

### Docker Compose (`docker-compose.yml`)

- Integration service command: `["bun", "test"]` -> `["npx", "vitest", "run"]`

### Source Files

- `index.ts`: Remove `.ts` extension from import path (`./gen/notifications/v1/service_pb.ts` -> `./gen/notifications/v1/service_pb`)
- `config.ts`, `types.ts`, `gen/`: No changes needed

## Out of Scope

- Upgrading any non-Bun dependencies
- Changing test logic or adding new tests
- Modifying the Go notification server
