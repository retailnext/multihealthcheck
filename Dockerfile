FROM golang as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src

COPY go.mod go.sum /src/

RUN go mod download \
&& go mod verify

COPY checker /src/checker
COPY handler /src/handler
COPY multihealthcheck.go /src/

RUN go build -a -v -trimpath -o multihealthcheck -ldflags="-s -w"

FROM gcr.io/distroless/static

COPY --from=builder /src/multihealthcheck .

ENTRYPOINT [ "/multihealthcheck" ]
