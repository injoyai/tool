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
ADD ./website   $WORKDIR/website
RUN chmod +x $WORKDIR/website

# 添加I18N多语言文件、静态文件、配置文件、模板文件
ADD config   $WORKDIR/config

###############################################################################
#                                   START
###############################################################################
WORKDIR $WORKDIR
CMD $WORKDIR/website
