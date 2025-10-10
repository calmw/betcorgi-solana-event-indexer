FROM golang AS builder
WORKDIR /app
ADD . /app
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
RUN go mod tidy
RUN go build -o /solana_indexer cmd/indexer/main.go

FROM golang
WORKDIR /app
COPY --from=builder /solana_indexer /
RUN chmod +x /solana_indexer

ENTRYPOINT ["/solana_indexer"]