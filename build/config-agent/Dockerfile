ARG GOLANG_BUILD_VERSION=stable

FROM onosproject/golang-build:$GOLANG_BUILD_VERSION
ENV GO111MODULE=on
COPY . /root/config-agent
WORKDIR /root/config-agent
RUN go mod download

ENTRYPOINT ["go", "run", "github.com/onosproject/config-models/cmd/config-agent"]
