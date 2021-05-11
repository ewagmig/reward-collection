FROM alpine:3.6
#更新Alpine的软件源
RUN echo "http://mirrors.aliyun.com/alpine/v3.9/community" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.9/main" >> /etc/apk/repositories

#安装vim bash curl wget net-tools
RUN apk update \
        && apk upgrade \
        && apk add --no-cache bash \
        bash-doc \
        bash-completion \
        && rm -rf /var/cache/apk/* \
        && /bin/bash

RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
    && wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.32-r0/glibc-2.32-r0.apk \
    && apk add glibc-2.32-r0.apk


# Overwrite the path to pass configuration or API index.html to server
ENV FABRIC_BAAS_CFG_PATH /etc/common/server

ENV BAAS_APPLICATION_HOMEDIR /var/common/server

# All data or files generated in runtime will be stored
# under the path /var/baas/server
RUN mkdir -p $FABRIC_BAAS_CFG_PATH/api $BAAS_APPLICATION_HOMEDIR \
$BAAS_APPLICATION_HOMEDIR/bin $BAAS_APPLICATION_HOMEDIR/config

# Copy the baasBackend server binary
COPY payload/common-backend /usr/local/bin
RUN chmod +x /usr/local/bin/common-backend

# Copy the BaaS configuration and API index.html
COPY payload/common-backend.yaml $FABRIC_BAAS_CFG_PATH
#COPY payload/index.html $FABRIC_BAAS_CFG_PATH/api

# Copy the docker entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT [ "docker-entrypoint.sh" ]

WORKDIR /var/common/server
EXPOSE 8005
CMD ["commonRun"]

LABEL baas.server.version=1.0.0
