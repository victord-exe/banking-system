# ==============================================================================
# STAGE 1: Dependencies
# ==============================================================================
FROM node:20-alpine AS dependencies

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# ==============================================================================
# STAGE 2: Development/Production
# ==============================================================================
FROM node:20-alpine

WORKDIR /app

# Copy dependencies from previous stage
COPY --from=dependencies /app/node_modules ./node_modules

# Copy source code
COPY . .

# Expose Vite dev server port
EXPOSE 5173

# Run development server with explicit host binding
# The --host flag ensures Vite binds to 0.0.0.0, allowing Docker to access it
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]