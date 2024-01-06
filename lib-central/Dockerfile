FROM golang:1.19 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o central-lib

FROM scratch

COPY --from=builder /app/central-lib /central-lib

EXPOSE 8080

ENV PORT=:8080

ENTRYPOINT ["/central-lib"]