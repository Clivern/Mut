FROM golang:1.27.3 AS builder

ARG MUT_VERSION=0.1.0
ARG MUT_COMMIT=none
ARG MUT_BUILD_DATE=unknown
ARG MUT_BUILT_BY=docker

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${MUT_VERSION} -X main.commit=${MUT_COMMIT} -X main.date=${MUT_BUILD_DATE} -X main.builtBy=${MUT_BUILT_BY}" \
    -o /out/mut .

RUN mkdir -p /out/configs /out/var/logs && \
    cp config.dist.yml /out/configs/config.dist.yml

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

COPY --from=builder --chown=65532:65532 /out/mut /app/mut
COPY --from=builder --chown=65532:65532 /out/configs /app/configs
COPY --from=builder --chown=65532:65532 /out/var /app/var

EXPOSE 8000

VOLUME ["/app/configs", "/app/var"]

USER 65532:65532

ENTRYPOINT ["/app/mut"]
CMD ["server", "-c", "/app/configs/config.dist.yml"]
