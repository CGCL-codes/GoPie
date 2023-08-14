FROM golang:1.19
RUN go env -w GOPROXY=https://goproxy.cn,direct && go get go.uber.org/goleak
RUN cd /tool && chmod +x ./script/patch.sh && ./script/patch.sh
WORKDIR /tool


