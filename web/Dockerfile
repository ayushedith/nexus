FROM node:18-alpine AS web-build
WORKDIR /app
COPY package*.json ./
RUN npm ci --silent
COPY . .
RUN npm run build

FROM node:18-alpine AS web-runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=web-build /app/ .
EXPOSE 3000
CMD ["npm", "run", "start"]
