FROM golang:1.14.3 AS builder
WORKDIR /go/src/github.com/bcspragu/Gobots

COPY go.mod go.sum *.go ./
COPY botapi/ botapi/
COPY engine/ engine/
COPY game/ game/

# Build the binary
RUN go get -u ./...
RUN go get -u github.com/gopherjs/gopherjs
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gobots github.com/bcspragu/Gobots

# Install Go 1.12.17 for GopherJS
RUN go get golang.org/dl/go1.12.17
RUN go1.12.17 download
COPY gopherjs/ gopherjs/
RUN GO111MODULE=off go1.12.17 get golang.org/x/net/context
RUN GO111MODULE=off go1.12.17 get zombiezen.com/go/capnproto2
RUN GOPHERJS_GOROOT="$(go1.12.17 env GOROOT)" gopherjs build github.com/bcspragu/Gobots/gopherjs --output=js/gopher.js 

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/bcspragu/Gobots/gobots /root/
COPY css/ css/
COPY img/ img/
COPY js/ js/
COPY templates/ templates/
CMD ["./gobots"]  
