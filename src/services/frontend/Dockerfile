FROM node:20-alpine AS builder
WORKDIR /app

# Install pnpm 9 (compatible with Node.js 20)
RUN npm install -g pnpm@9

# Configure pnpm to use hoisted node_modules (npm-compatible)
RUN echo "node-linker=hoisted" > .npmrc

# @hlv:sec [SECRET_HANDLING] — Keycloak OIDC config passed via build args, baked into static bundle
ARG VITE_KEYCLOAK_REALM=vedo-core
ARG VITE_KEYCLOAK_CLIENT_ID=vedo-spa
ARG VITE_KEYCLOAK_URL=http://127.0.0.1:8443

ENV VITE_KEYCLOAK_REALM=${VITE_KEYCLOAK_REALM}
ENV VITE_KEYCLOAK_CLIENT_ID=${VITE_KEYCLOAK_CLIENT_ID}
ENV VITE_KEYCLOAK_URL=${VITE_KEYCLOAK_URL}

# Copy dependency manifests
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile --ignore-scripts

# Copy source code
COPY . .

# Build — vite handles TS transpilation without typecheck
RUN node --input-type=module -e "import { build } from 'vite'; build()"

# Production stage
FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
ENTRYPOINT ["nginx", "-g", "daemon off;"]
