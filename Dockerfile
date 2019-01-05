FROM alpine:3.8
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app
WORKDIR /home/app
COPY ./webhook-bridge .
USER app
HEALTHCHECK --interval=2s CMD [ -e ./.running ] || exit 1
CMD ["./webhook-bridge"]