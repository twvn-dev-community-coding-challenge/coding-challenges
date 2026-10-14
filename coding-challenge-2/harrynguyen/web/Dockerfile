# syntax=docker/dockerfile:1
# Production SPA: Vite build → nginx (static + /api reverse proxy for same-origin fetch).
FROM node:22-alpine AS build
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@9 --activate

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

# Optional: set at build time if the API is on another origin (e.g. -e VITE_API_BASE=https://api.example.com).
ARG VITE_API_BASE=
ENV VITE_API_BASE=$VITE_API_BASE

RUN pnpm run build

FROM nginx:1.27-alpine AS runtime
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
