## Build the app
FROM node:23-alpine as builder

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY tsconfig.json .
COPY src/ ./src/
RUN npm run build

## Run the app
FROM node:23-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /app/dist ./dist
RUN mkdir -p /app/data
VOLUME /app/data
ENV NODE_ENV=production
ENV WATCH_DIRECTORY=/app/data

CMD ["node", "dist/main.js"]
