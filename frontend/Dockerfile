FROM oven/bun:1-alpine AS builder

ARG VITE_API_URL
ARG VITE_DEFAULT_REDIRECT_URL
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_DEFAULT_REDIRECT_URL=${VITE_DEFAULT_REDIRECT_URL}

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

RUN bun run build

FROM oven/bun:1-alpine AS runner

WORKDIR /app

COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json .

EXPOSE 3000

CMD ["bun", "run", "start"]
