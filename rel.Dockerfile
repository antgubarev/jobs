FROM alpine

ADD dist/server_linux_amd64/server /server

EXPOSE 80

ENTRYPOINT ["/server"]