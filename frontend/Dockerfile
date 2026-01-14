# Multi-stage build using Bun
FROM oven/bun:latest AS builder
WORKDIR /app

# Install deps and build
COPY package.json bun.lock* ./
COPY . .
RUN bun install
RUN bun run build

# Final image: only runtime + built assets
FROM oven/bun:latest AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 4000
CMD ["bun", "run", "preview", "--", "--port", "4000", "--host", "0.0.0.0"]
