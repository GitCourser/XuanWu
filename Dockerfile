# 编译
FROM golang:1.24-alpine as builder
WORKDIR /app/
COPY ./ ./
RUN apk add --no-cache bash curl gcc git musl-dev && \
    go build -o ./build/xuanwu -ldflags="-w -s" .

# 获取nodejs
FROM node:20-alpine AS nodejs

# 最终镜像
FROM python:3.11-alpine

WORKDIR /app/
COPY --from=nodejs /usr/local/lib/node_modules/. /usr/local/lib/node_modules/
COPY --from=nodejs /usr/local/bin/. /usr/local/bin/
COPY --from=builder /app/build/. .

ENV XW_HOME=/app
ENV TZ=Asia/Shanghai
ENV PYTHONPATH=${XW_HOME}
ENV NODE_PATH="/usr/local/lib/node_modules:${XW_HOME}"
ENV NODE_OPTIONS=--tls-cipher-list=DEFAULT@SECLEVEL=0

RUN apk add --no-cache libstdc++ libgcc && \
    npm install -g --no-cache got@~11.8.0 && \
    npm config set registry https://registry.npmmirror.com && \
    pip config set global.no-cache-dir true && \
    pip install requests && \
    pip config set global.index-url https://mirrors.huaweicloud.com/artifactory/pypi-public/simple && \
    cp -a /etc/apk/repositories /etc/apk/repositories.bak && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.huaweicloud.com/g' /etc/apk/repositories && \
    wget https://github.com/whyour/qinglong/raw/refs/heads/develop/sample/notify.js -O /app/notify.js && \
    wget https://github.com/whyour/qinglong/raw/refs/heads/develop/sample/notify.py -O /app/notify.py && \
    sed -i 's/+ (await one(/\/\/ (await one(/' /app/notify.js && \
    sed -i 's/+ one(/# one(/' /app/notify.py && \
    printf '#!/bin/sh\n\n. ${XW_HOME}/data/Env.sh\n"$@"\n' > /usr/local/bin/xw && \
    chmod +x /usr/local/bin/xw && \
    chmod +x entrypoint.sh && \
    rm -f /usr/local/bin/yarn* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/* && \
    rm -rf /tmp/*

ENTRYPOINT ["./entrypoint.sh"]