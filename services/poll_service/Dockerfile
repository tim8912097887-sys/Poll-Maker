ARG NODE_VERSION=24
FROM node:${NODE_VERSION}-alpine AS base

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV PNPM_CONFIG_STORE_DIR="/pnpm/store"
RUN corepack enable

# ==========================================
# 1. Prereq Stage: Copy all structural files
# ==========================================
FROM base AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Install ALL dependencies (including devDeps for building)
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

# Copy the rest of the source code (Ensure .dockerignore ignores local node_modules!)
COPY . .

RUN pnpm build

# ==========================================
# 3. Final Lean Production Stage
# ==========================================
FROM node:${NODE_VERSION}-alpine AS production
WORKDIR /app

# Safely switch to non-root user
USER node

# Copy the isolated self-contained application from the builder stage
# This includes the exact production node_modules and the exact production build!
COPY --from=builder --chown=node:node /app/dist ./dist
COPY --from=builder --chown=node:node /app/node_modules ./node_modules
COPY --chown=node:nod drizzle ./drizzle

EXPOSE 3000

# Double check your entrypoint path inside the build artifact folder
CMD ["node", "dist/src/main.js"]