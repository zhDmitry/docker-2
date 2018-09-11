FROM golang
WORKDIR /go/src/test
COPY . .
RUN go get -d -v .
EXPOSE 1234
CMD go run .