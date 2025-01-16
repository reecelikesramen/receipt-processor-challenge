FROM golang:1.23.4

WORKDIR /usr/src/app

# only redownload dependencies if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# test
RUN go test -test.v

# build
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]