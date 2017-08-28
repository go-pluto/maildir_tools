FROM alpine:3.6

ADD ./maildir_exporter /bin/
RUN apk add -U ca-certificates

ENTRYPOINT ["/bin/maildir_exporter"]
CMD ["-logLevel", "debug", "-maildirRootPath", "/data/maildirs"]
