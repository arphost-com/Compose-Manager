FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

FROM alpine:3.22
RUN apk add --no-cache nginx && \
    apk upgrade --no-cache && \
    mkdir -p /run/nginx /var/lib/nginx/tmp/client_body /usr/share/nginx/html && \
    chown -R nginx:nginx /run/nginx /var/lib/nginx /usr/share/nginx/html /etc/nginx/http.d
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/http.d/default.conf
USER nginx
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/ || exit 1
ENTRYPOINT []
CMD ["nginx", "-g", "daemon off;"]
