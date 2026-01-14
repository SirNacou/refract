FROM oven/bun:alpine AS builder

WORKDIR /app

COPY package.json bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

COPY --from=root .env .

RUN bun run build

FROM oven/bun:alpine

WORKDIR /app

COPY --from=builder /app/.output ./

CMD ["bun", "./server/index.mjs"]