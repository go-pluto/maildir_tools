FROM alpine:3.6

ADD ./maildir_exporter /bin/

ENTRYPOINT ["/bin/maildir_exporter"]
CMD ["-logLevel", "debug", "-maildirRootPath", "/data/maildirs"]
