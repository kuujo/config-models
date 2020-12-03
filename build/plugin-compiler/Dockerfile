ARG GOLANG_BUILD_VERSION=stable

FROM onosproject/golang-build:$GOLANG_BUILD_VERSION
ENV GO111MODULE=on
COPY . /root/plugin-compiler
WORKDIR /root/plugin-compiler

ENTRYPOINT ["go", "run", "github.com/onosproject/config-models/cmd/config-model"]
