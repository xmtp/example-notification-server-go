FROM node:22-slim AS base
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
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
