# syntax=docker/dockerfile:1

FROM node:20-bookworm-slim AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN rm -rf internal/webui/dist
COPY --from=web-build /src/web/dist ./internal/webui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/meowbridge ./cmd/meowbridge
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /data
ENV HTTP_PORT=8080
COPY --from=go-build --chown=nonroot:nonroot /out/meowbridge /usr/local/bin/meowbridge
COPY --from=go-build --chown=nonroot:nonroot /out/data /data
EXPOSE 8080
VOLUME ["/data"]
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/meowbridge"]
