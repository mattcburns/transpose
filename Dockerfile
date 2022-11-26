FROM golang:latest as build
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o transpose

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=build /build/transpose /transpose
ENTRYPOINT [ "/transpose" ]