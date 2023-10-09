FROM golang:1.17

RUN mkdir /forum

ADD . /forum

WORKDIR /forum

RUN go build  -o main .

EXPOSE 8080

CMD ["/forum/main"]
