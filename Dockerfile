
FROM golang:1.26-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY golang/go.mod golang/go.sum ./
RUN go mod download
COPY golang/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/keights-pod-tracker .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/keights-pod-tracker /keights-pod-tracker
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/keights-pod-tracker"]
