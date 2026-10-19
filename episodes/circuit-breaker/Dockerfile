FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /demo .

FROM scratch
COPY --from=build /demo /demo
USER 65532:65532
ENTRYPOINT ["/demo"]
