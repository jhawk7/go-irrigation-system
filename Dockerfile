FROM golang:1.20.8-alpine3.18 AS build
WORKDIR /builder
COPY . ./
RUN go mod download
RUN GOOS=linux GOARCH=arm go build cmd/app/main.go

FROM golang:1.20.8-alpine3.18
WORKDIR /app
COPY --from=build /builder/main .
CMD ["./main"]
# i2c bus folder must be mounted from pi