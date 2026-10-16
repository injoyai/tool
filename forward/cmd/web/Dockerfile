FROM alpine:latest

LABEL maintainer="injoy"

###############################################################################
#                                INSTALLATION
###############################################################################

# 设置固定的项目路径
ENV WORKDIR /root
ENV TZ=Asia/Shanghai

# 安装时区数据
RUN apk add --no-cache tzdata

# 添加应用可执行文件，并设置执行权限
ADD ./forward   $WORKDIR/forward
RUN chmod +x $WORKDIR/forward

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
CMD $WORKDIR/forward
