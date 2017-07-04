#!/bin/bash -eu

pushd $(dirname $0)/..
  export GOPATH=$PWD
  export PATH=$GOPATH/bin:$PATH

  pushd src/github.com/pivotal-cf-experimental/kafka-example-service-adapter
    go get -v github.com/tools/godep
    godep restore

    go install -v github.com/onsi/ginkgo/ginkgo

    ginkgo -r -randomizeAllSpecs -randomizeSuites -race -keepGoing -failOnPending -cover
  popd
popd
