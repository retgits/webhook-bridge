FROM alpine:3.9
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app
WORKDIR /home/app
COPY ./pubnub .
USER app
HEALTHCHECK --interval=2s CMD [ -e ./.running ] || exit 1
CMD ["./pubnub"]