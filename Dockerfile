ARG GO_VERSION=1.23
ARG TARGETARCH
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    # --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -ldflags='-s -w -extldflags "-static"' -o /bin/vote ./main.go

FROM alpine:latest AS final

WORKDIR /app
COPY --from=build /bin/vote /vote

EXPOSE 3000

CMD [ "/vote"]
