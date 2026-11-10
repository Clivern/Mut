FROM golang:1.27.4 AS builder

ARG MUT_AGENT_VERSION=0.1.0
ARG MUT_AGENT_COMMIT=none
ARG MUT_AGENT_BUILD_DATE=unknown
ARG MUT_AGENT_BUILT_BY=docker

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${MUT_AGENT_VERSION} -X main.commit=${MUT_AGENT_COMMIT} -X main.date=${MUT_AGENT_BUILD_DATE} -X main.builtBy=${MUT_AGENT_BUILT_BY}" \
    -o /out/mut-agent .

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /out/mut-agent /mut-agent

USER 65532:65532

ENTRYPOINT ["/mut-agent"]
CMD ["run"]
