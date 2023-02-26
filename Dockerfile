# Stage 2: Build the binary
FROM golang:alpine AS binary-builder
RUN apk update && apk upgrade && apk --update add git
WORKDIR /builder
COPY go.mod go.sum ./
RUN go mod download
ENV ENV=production
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags='-w -s -extldflags "-static"' -a \
  -o tkd

# Stage 3: Run the binary
FROM busybox AS busybox


ENV APP_DIRECTORY=/app
ENV TABLE_LOG=job_queues
ENV FILE_REPORT_LOCATION=/app/assets/reports

ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379
ENV REDIS_DATABASE=0
ENV REDIS_USERNAME=sinotif
ENV REDIS_PASSWORD=SinotifDev

ENV DB_USER=tkd
ENV DB_PASSWORD=j0a0w3Ed91VMEomJ5F
ENV DB_NAME=assessment
ENV DB_HOST=202.157.177.80
ENV DB_PORT=5431

WORKDIR /app
RUN mkdir -p /app/assets/report
COPY --from=binary-builder --chown=nonroot:nonroot /builder/tkd /app
ENTRYPOINT ["/app/tkd"]

