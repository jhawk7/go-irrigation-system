FROM golang:1.20 AS build
WORKDIR /go/cmd/github.com/jhawk7/go-pi-irrigation/
COPY . ./
RUN go mod download
RUN GOOS=linux GOARCH=arm go build -o bin/irrigation

FROM golang:1.20
WORKDIR /
COPY --from=build /go/cmd/github.com/jhawk7/go-pi-irrigation/bin/irrigation ./
EXPOSE 8080
CMD ["./irrigation"]
# i2c bus folder must be mounted from pi