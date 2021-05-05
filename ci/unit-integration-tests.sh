#!/bin/bash -eu

pushd $(dirname $0)/..
  export GOPATH=$PWD
  export PATH=$GOPATH/bin:$PATH

  pushd src/github.com/pivotal-cf-experimental/kafka-example-service-adapter
    go get -u github.com/onsi/ginkgo/ginkgo

    GO111MODULE=off ginkgo -r -randomizeAllSpecs -randomizeSuites -race -keepGoing -failOnPending -cover
  popd
popd
