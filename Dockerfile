FROM golang
WORKDIR /go/src/app
COPY . .
RUN go get -v .
EXPOSE 1234
RUN go build .
CMD ./app -c $CONNECTION_STR