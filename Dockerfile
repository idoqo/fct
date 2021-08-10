FROM golang:alpine AS builder
MAINTAINER Michael Okoko<hi@mchl.xyz>

COPY go.mod go.sum /go/src/gitlab.com/idoko/flatcartag/
WORKDIR /go/src/gitlab.com/idoko/flatcartag
RUN go mod download
COPY . /go/src/gitlab.com/idoko/flatcartag
RUN CGO_ENABLED=0 GOOS=linux go build -a --installsuffix cgo -o build/fct ./cmd/flatcartag

FROM alpine
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY --from=builder /go/src/gitlab.com/idoko/flatcartag/build/fct /usr/bin/fct

ENTRYPOINT ["/usr/bin/fct"]