FROM oven/bun:1-alpine AS builder

ARG VITE_APP_TITLE
ARG VITE_API_URL
ARG VITE_PORT
ENV VITE_APP_TITLE=${VITE_APP_TITLE}
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_PORT=${VITE_PORT}

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

ARG VITE_PORT

CMD [ "sh", "-c", "PORT=${VITE_PORT:-3000} bun --bun run start" ]

LABEL org.opencontainers.image.source=https://github.com/SirNacou/refract
