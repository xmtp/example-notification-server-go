FROM oven/bun:1 AS base
WORKDIR /usr/app

# FROM base AS install
# RUN mkdir -p /temp/dev
# COPY package.json bun.lockb /temp/dev/
# RUN cd /temp/dev && bun install --frozen-lockfile

FROM base
RUN bun install -g yarn
# COPY --from=install /temp/dev/node_modules node_modules
COPY . .
RUN yarn install; bun install

USER bun
EXPOSE 7777/tcp

# Add these debug commands before the entrypoint
RUN ls -la /usr/app/node_modules/@xmtp
RUN find /usr/app/node_modules -name "*linux-x64-gnu*"

ENTRYPOINT [ "bun", "test" ]
