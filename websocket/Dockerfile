#-----------------------------------------------------------------------------
FROM golang:alpine AS builder

RUN apk add --no-cache upx

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /go/src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# RUN ls
RUN go build \
      -trimpath \
      -ldflags="-s -w -extldflags '-static'" \
      -o /go/bin/main \
	  .

RUN upx --lzma /go/bin/main

#-----------------------------------------------------------------------------
FROM scratch

ENV GODEBUG=madvdontneed=1

COPY --from=builder /go/bin/main .

EXPOSE 9090
ENTRYPOINT ["./main"]
