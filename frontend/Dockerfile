# ---- deps ----
FROM node:20-alpine AS deps

WORKDIR /app
COPY package*.json ./
RUN npm ci

# ---- dev target (used by docker-compose.yml) ----
FROM deps AS dev

COPY . .
EXPOSE 5173
# --host exposes the vite dev server outside the container
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# ---- build ----
FROM deps AS builder

COPY . .
RUN npm run build

# ---- production target (used by docker-compose.prod.yml) ----
FROM nginx:1.27-alpine AS production

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
