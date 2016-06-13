FROM golang:1.6 
RUN wget https://github.com/Masterminds/glide/releases/download/0.10.2/glide-0.10.2-linux-amd64.tar.gz && \
    tar xvfz glide-0.10.2-linux-amd64.tar.gz -C /usr/local/bin --strip-components=1 linux-amd64/glide && \
    rm glide-0.10.2-linux-amd64.tar.gz
RUN mkdir -p /go/src/github.com/cluda/btcdata
ADD . /go/src/github.com/cluda/btcdata
WORKDIR /go/src/github.com/cluda/btcdata 
RUN glide install
RUN go run main.go