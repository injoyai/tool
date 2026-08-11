#!/bin/bash
set -e

# ===== 镜像配置(按需修改) =====
REGISTRY="crpi-ayrx20sj8nkmrgmh.cn-hangzhou.personal.cr.aliyuncs.com"              # 例: registry.cn-hangzhou.aliyuncs.com  留空=不推送,仅本地构建
NAMESPACE="injoyai"      # 命名空间/用户名
IMAGE_NAME="timer"       # 镜像名
VERSION="latest"             # 版本号(可用: git describe --tags --always)
PLATFORMS="linux/amd64"             # 多架构,留空=当前架构; 例: linux/amd64,linux/arm64
# ==============================

cd "$(dirname "$0")"

FULL_NAME="${REGISTRY:+$REGISTRY/}${NAMESPACE}/${IMAGE_NAME}"

echo "==> 构建镜像: $FULL_NAME:$VERSION"

if [ -n "$PLATFORMS" ]; then
    # 多架构构建(需 docker buildx,且 --push 直接推送,不能只本地 load)
    docker buildx build \
        --platform "$PLATFORMS" \
        -t "$FULL_NAME:$VERSION" \
        -t "$FULL_NAME:latest" \
        --push .
else
    # 单架构构建
    docker build \
        -t "$FULL_NAME:$VERSION" \
        -t "$FULL_NAME:latest" \
        .
fi

if [ -z "$REGISTRY" ]; then
    echo "==> 未配置 REGISTRY,仅本地构建完成: $FULL_NAME:$VERSION"
    echo "    如需推送,请设置 REGISTRY 变量后重新运行"
    exit 0
fi

if [ -z "$PLATFORMS" ]; then
    echo "==> 推送镜像: $FULL_NAME:$VERSION"
    docker push "$FULL_NAME:$VERSION"
    docker push "$FULL_NAME:latest"
fi

echo "==> 完成: $FULL_NAME:$VERSION"
