# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy source code
COPY . .

# Install production dependencies only
RUN yarn install


# Run the application
CMD ["yarn", "dev"]
