FROM golang:1.22.2
COPY . /go/src/github.com/quilla-hq/quilla
WORKDIR /go/src/github.com/quilla-hq/quilla
RUN make install

FROM node:18.17.1-alpine
WORKDIR /app
COPY ui /app
RUN yarn
RUN yarn run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

VOLUME /data
ENV XDG_DATA_HOME /data

COPY --from=0 /go/bin/quilla /bin/quilla
COPY --from=1 /app/build /www
ENTRYPOINT ["/bin/quilla"]
EXPOSE 9300
