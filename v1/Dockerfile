ARG GO_VERSION=1.25.6
ARG ALPINE_VERSION=3.20
ARG DISTROLESS_TAG=nonroot

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
        -trimpath \
        -mod=readonly \
        -o /out/server \
        ./dev/cmd/server


FROM gcr.io/distroless/static-debian12:${DISTROLESS_TAG} AS runtime

COPY --from=builder --chown=nonroot:nonroot --chmod=0555 /out/server /server

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/server"]
