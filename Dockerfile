FROM golang:alpine AS isbpbuildmaster
WORKDIR /go/src/github.com/nikitamirzani323/BTANGKAS_CLIENT_API
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .


# Moving the binary to the 'final Image' to make it smaller
FROM alpine:latest as totosvelterelease
WORKDIR /app
RUN apk add tzdata
RUN mkdir -p ./frontend/public
COPY --from=isbpbuildmaster /go/src/github.com/nikitamirzani323/BTANGKAS_CLIENT_API/app .
COPY --from=isbpbuildmaster /go/src/github.com/nikitamirzani323/BTANGKAS_CLIENT_API/env-sample /app/.env

ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

EXPOSE 1117
CMD ["./app"]