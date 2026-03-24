FROM oven/bun:slim AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts

COPY tsconfig.json ./
COPY src/ src/
RUN bun run build

FROM oven/bun:slim

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production --ignore-scripts

COPY --from=builder /app/dist/ dist/

ENV PORT=8080
EXPOSE 8080

USER bun
CMD ["bun", "run", "start"]
