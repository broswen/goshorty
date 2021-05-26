FROM golang

WORKDIR /go/src/app

COPY ./go.mod .

RUN go get -d -v ./...

COPY ./ .

RUN go build 

EXPOSE 8080

CMD ["./goshort"]



