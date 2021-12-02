# Copyright 2020 Coinbase, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.17  as golang-builder

ARG ERIGON_TAG=v2021.12.01

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

# Compile Erigon
# VERSION: go-ethereum v.1.10.8
RUN git clone --recurse-submodules -j8 https://github.com/ledgerwatch/erigon.git ./erigon-node \
  && cd erigon-node \
  && git checkout ${ERIGON_TAG} \
  && make erigon

# Compile rosetta-ethereum
# Use native remote build context to build in any directory
COPY . src
RUN cd src \
  && go build

RUN mv src/rosetta-ethereum /app/rosetta-ethereum \
  && mkdir /app/ethereum \
  && mv src/ethereum/call_tracer.js /app/ethereum/call_tracer.js \
  && mv src/ethereum/geth.toml /app/ethereum/geth.toml \
  && rm -rf src

## Build Final Image
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from geth-builder
COPY --from=golang-builder /app/erigon-node/build/bin/erigon /app/erigon

# Copy binary from rosetta-builder
COPY --from=golang-builder /app/ethereum /app/ethereum
COPY --from=golang-builder /app/rosetta-ethereum /app/rosetta-ethereum

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

CMD ["/app/rosetta-ethereum", "run"]
