FROM node:8.16-alpine as builder

WORKDIR /app

COPY ./ui/package.json /app/package.json
COPY ./ui/package-lock.json /app/package-lock.json
RUN npm install

COPY ./ui /app

RUN npm run build

FROM abiosoft/caddy:1.0.0

COPY ./Caddyfile /etc/Caddyfile

COPY --from=builder /app/build /srv