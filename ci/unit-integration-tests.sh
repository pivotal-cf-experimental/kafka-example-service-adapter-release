#!/bin/bash -eu

abspath() {
  cd "$1"
  pwd
}

cd $(dirname $0)/..
export GOPATH=$PWD

cd src/github.com/pivotal-cf-experimental/kafka-example-service-adapter
ginkgo -r -randomizeAllSpecs -randomizeSuites -race -keepGoing -failOnPending -cover
