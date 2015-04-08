FROM golang
MAINTAINER Constantin Schomburg <me@cschomburg.com>

WORKDIR /go/src/github.com/xconstruct/stark
RUN mkdir -p /go/src/github.com/xconstruct/stark
COPY . /go/src/github.com/xconstruct/stark

RUN go get ./cmd/starkd && go get ./cmd/tars
RUN go install ./cmd/starkd && go install ./cmd/tars

RUN useradd -m stark -d /stark
USER stark
WORKDIR /stark

COPY ./assets /stark/assets
COPY ./assets/luascripts /stark/luascripts
RUN echo '{}' > /stark/server.json
RUN echo '{}' > /stark/client.json

VOLUME /stark
CMD ["starkd"]

EXPOSE 5000
EXPOSE 23100
EXPOSE 23443
