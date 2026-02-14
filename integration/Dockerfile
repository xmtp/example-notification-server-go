FROM oven/bun:1 as base
WORKDIR /usr/app

# install dependencies into temp directory
# this will cache them and speed up future builds
FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
RUN cd /temp/dev && bun install --frozen-lockfile

# The real release artifact
FROM base
COPY --from=install /temp/dev/node_modules node_modules
COPY . .

# run the app
USER bun
EXPOSE 7777/tcp
ENTRYPOINT [ "bun", "test" ]