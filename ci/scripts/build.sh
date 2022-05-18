#!/bin/bash -eux

pushd dp-interactives-api
  make build
  cp build/dp-interactives-api Dockerfile.concourse ../build
popd
