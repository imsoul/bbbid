FROM alpine
MAINTAINER chenzhiyin    "1981330085@qq.com"

#RUN sed -i 's!http://dl-cdn.alpinelinux.org/!https://mirrors.ustc.edu.cn/!g' /etc/apk/repositories
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
WORKDIR /opt/server

RUN apk update --no-cache
RUN apk add --no-cache ca-certificates && \
    update-ca-certificates
RUN apk add --no-cache tzdata
ENV TZ Asia/Shanghai

COPY ./bin/bbbid ./bbbid
RUN chmod +x ./bbbid
COPY ./configs ./configs

EXPOSE 8810
EXPOSE 8811
VOLUME /opt/server/configs
VOLUME /opt/server/logs
CMD ["./bbbid", "-conf", "/opt/server/configs"]