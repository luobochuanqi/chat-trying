FROM --platform=$TARGETPLATFORM golang:1.20-alpine AS backend

WORKDIR /backend
COPY . .

ARG TARGETARCH
ARG TARGETOS
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,direct

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && apk add --no-cache gcc musl-dev g++ make linux-headers

RUN go build -o chat -a -ldflags="-extldflags=-static" .

FROM node:18 AS frontend

WORKDIR /app
COPY ./app .

RUN npm install -g pnpm --registry=https://registry.npmmirror.com && \
    pnpm config set registry https://registry.npmmirror.com && \
    pnpm install && pnpm run build && rm -rf node_modules src

FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk upgrade --no-cache && apk add --no-cache wget ca-certificates tzdata && \
    update-ca-certificates 2>/dev/null || true

RUN echo "Asia/Shanghai" > /etc/timezone && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

COPY --from=backend /backend/chat /chat
COPY --from=backend /backend/config.example.yaml /config.example.yaml
COPY --from=backend /backend/utils/templates /utils/templates
COPY --from=backend /backend/addition/article/template.docx /addition/article/template.docx
COPY --from=frontend /app/dist /app/dist

VOLUME ["/config", "/logs", "/storage"]

EXPOSE 8094

CMD ["./chat"]
