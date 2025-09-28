# Dockerfile.rules (Node/TS)
FROM node:20-alpine AS deps
WORKDIR /app
COPY ./package*.json ./
RUN npm install --package-lock-only
RUN npm ci

FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY ./ ./
RUN npm install --save-dev @types/json-logic-js
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production PORT=3000
COPY --from=build /app/dist ./dist
COPY --from=deps /app/node_modules ./node_modules
EXPOSE 3000
CMD ["node", "dist/index.js"]
