# builder
FROM golang:1.13-alpine as builder

RUN apk add --update-cache gcc libc-dev
WORKDIR /app
COPY . .
RUN go build -o koubachi-goserver .

# runner
FROM alpine:latest

ARG UID=1000
ARG GID=1000

WORKDIR /app
RUN apk add --update --no-cache ca-certificates dumb-init
COPY --from=builder /app/koubachi-goserver .
COPY ./assets ./assets

# setup environment
ENV PATH "/app:${PATH}"
RUN addgroup --gid $GID -S koubachi && \
    adduser --uid ${UID} -S -G koubachi koubachi && \
    chown -R koubachi:koubachi /app
USER koubachi
EXPOSE 8005

ENTRYPOINT ["dumb-init", "--"]
CMD ["koubachi-goserver"]
