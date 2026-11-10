FROM docker.io/library/golang:1.27.3-alpine3.24 AS build

WORKDIR /src
COPY . .

RUN go test ./...
RUN CGO_ENABLED=0 go build -trimpath -o /out/map_generator ./cmd/map_generator

FROM scratch

COPY --from=build /out/map_generator /map_generator

WORKDIR /data

ENTRYPOINT ["/map_generator"]