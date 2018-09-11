FROM golang

ADD . /go/src/trlogic
#ADD ./model /go/src/github.com/vvizard75/trlogic/model
#ADD ./html /go/src/github.com/vvizard75/trlogic/html

COPY ./html /go/html




RUN go install /go/src/trlogic

#ENTRYPOINT pwd
ENTRYPOINT /go/bin/trlogic

EXPOSE 8081