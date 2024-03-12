FROM regi.acloud.run/docker.io/library/alpine:3.18.4


# make sure root login is disabled
RUN sed -i -e 's/^root::/root:!:/' /etc/shadow && \
    apk update && apk upgrade && rm -Rf /var/cache/apk/*

# 작업 디렉토리를 설정합니다.
WORKDIR Goproxy

RUN chmod +x /Goproxy

# 현재 디렉토리의 파일을 도커에 복사합니다.
COPY bin/Goproxy /bin/Goproxy



# RUN go build -o main /repo/bin/Goproxy

EXPOSE 9999
CMD ["/bin/Goproxy"]