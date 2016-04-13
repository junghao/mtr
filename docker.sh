#!/bin/bash -e

# code will be compiled in this container
BUILD_CONTAINER=golang:1.6.1-alpine

DOCKER_TMP=docker-build-tmp

mkdir -p $DOCKER_TMP
chmod +s $DOCKER_TMP
mkdir -p ${DOCKER_TMP}/common/etc/ssl/certs
mkdir -p ${DOCKER_TMP}/common/usr/share

# Prefix for the logs
BUILD='-X github.com/GeoNet/mtr/vendor/github.com/GeoNet/log/logentries.Prefix=git-'`git rev-parse --short HEAD`

# Build all executables in the Golang container.  Output statically linked binaries to docker-build-tmp
# Assemble common resource for ssl and timezones
docker run -e "GOBIN=/usr/src/go/src/github.com/GeoNet/mtr/${DOCKER_TMP}" -e "GOPATH=/usr/src/go" -e "CGO_ENABLED=0" -e "GOOS=linux" -e "BUILD=$BUILD" --rm \
	-v "$PWD":/usr/src/go/src/github.com/GeoNet/mtr \
	-w /usr/src/go/src/github.com/GeoNet/mtr ${BUILD_CONTAINER} \
	go install -a  -ldflags "${BUILD}" -installsuffix cgo ./...; \
	cp /etc/ssl/certs/ca-certificates.crt ${DOCKER_TMP}/common/etc/ssl/certs; \
	cp -Ra /usr/share/zoneinfo ${DOCKER_TMP}/common/usr/share

# Assemble common resource for user.
echo "nobody:x:65534:65534:Nobody:/:" > ${DOCKER_TMP}/common/etc/passwd

# Docker images for web apps with an open port 
for i in mtr-api
do
	echo "FROM scratch" > docker-build-tmp/Dockerfile
	echo "ADD common ${i} /" >> docker-build-tmp/Dockerfile
	echo "USER nobody" >> docker-build-tmp/Dockerfile
	echo "EXPOSE 8080" >> docker-build-tmp/Dockerfile
	echo "CMD [\"/${i}\"]" >> docker-build-tmp/Dockerfile
	docker build --rm=true -t quay.io/geonet/geonet-web:$i -f docker-build-tmp/Dockerfile docker-build-tmp
done
