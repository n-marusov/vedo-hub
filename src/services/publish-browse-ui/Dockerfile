FROM node:20-alpine AS builder
WORKDIR /app

# Install pnpm 9 (compatible with Node.js 20)
RUN npm install -g pnpm@9

# Configure pnpm to use hoisted node_modules (npm-compatible)
RUN echo "node-linker=hoisted" > .npmrc

# Copy dependency manifests
COPY package.json ./

# Install dependencies
RUN pnpm install --ignore-scripts

# Copy source code
COPY . .

# Build — vite handles TS transpilation without typecheck
RUN node --input-type=module -e "import { build } from 'vite'; build()"

# Production stage
FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3002
ENTRYPOINT ["nginx", "-g", "daemon off;"]
