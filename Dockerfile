FROM golang:1.21.5-alpine AS builder
RUN apk update && apk add --no-cache git

WORKDIR /go/saveIMG
COPY ./saveIMG .

RUN go install
RUN go build -o /go/saveIMG/bin/saveIMG

FROM scratch
COPY --from=builder /go/saveIMG/bin/saveIMG .
CMD ["/saveIMG"]