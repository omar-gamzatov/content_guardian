# Stage 1: deps + build
FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./
# Устанавливаем зависимости (любой менеджер — npm/pnpm/yarn)
RUN npm ci || npm install
COPY tsconfig.json ./tsconfig.json
COPY public ./public
COPY src ./src
RUN npm run build

# Stage 2: runtime
FROM node:20-alpine
ENV NODE_ENV=production
WORKDIR /app
# Только нужные файлы
COPY --from=build /app/dist ./dist
COPY package.json ./
RUN npm install --omit=dev && npm cache clean --force
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD wget -q -T 3 -O - http://127.0.0.1:3000/healthz || exit 1
CMD ["node", "dist/server.js"]
