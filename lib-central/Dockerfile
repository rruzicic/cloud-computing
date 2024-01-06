FROM golang:1.19 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o lib-central

FROM scratch

COPY --from=builder /app/lib-central /lib-central

EXPOSE 8080

ENV PORT=:8080

ENTRYPOINT ["/lib-central"]