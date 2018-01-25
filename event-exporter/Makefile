# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: container

ENVVAR = GOOS=linux GOARCH=amd64 CGO_ENABLED=0
BINARY_NAME = event-exporter
GOPACKAGES=$(shell go list ./... | grep -v /vendor/) # With glide: GOPACKAGES=$(shell glide novendor)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*"| grep -v bindata.go | grep -v bindata_assetfs.go)

PREFIX = qrtp
IMAGE_NAME = k8s-stackdriver
TAG = ${TRAVIS_COMMIT}

deps:
	glide install --strip-vendor
	go get github.com/golang/lint/golint
	go get github.com/pierrre/gotestcover
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/rakyll/statik
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/elazarl/go-bindata-assetfs/...
	gometalinter --install

build:
	${ENVVAR} go build -a -o ${BINARY_NAME}

dofmt:
	go fmt ${GOPACKAGES}

fmt:
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Please run gofmt -l -w on your code.' && exit 1; fi

lint:
	$(GOPATH)/bin/golint -set_exit_status=true sinks/
	$(GOPATH)/bin/golint -set_exit_status=true watchers/
	$(GOPATH)/bin/golint -set_exit_status=true utils/
	$(GOPATH)/bin/golint -set_exit_status=true .

vet:
	gometalinter --disable-all --enable=vet --enable=golint --enable=vetshadow --enable=ineffassign --enable=goconst --tests  --vendor --exclude=sinks/activitytracker/client/pb/trail.pb.go ./...

test:
	go test ./...

coverage:
	./hack/coverage.sh

container: build
	docker build --pull -t ${PREFIX}/${IMAGE_NAME}:${TAG} .

push: container
	docker push ${PREFIX}/${IMAGE_NAME}:${TAG}

clean:
	rm -rf ${BINARY_NAME}
