#!/bin/bash

mkdir -p docker-build-tmp
chmod +s docker-build-tmp

# Build all executables in the golang-godep container.  Output statically linked binaries to docker-build-tmp
docker run -e "GOBIN=/usr/src/go/src/github.com/GeoNet/mtr/docker-build-tmp" -e "GOPATH=/usr/src/go" -e "CGO_ENABLED=0" -e "GOOS=linux" --rm -v \
"$PWD":/usr/src/go/src/github.com/GeoNet/mtr -w /usr/src/go/src/github.com/GeoNet/mtr golang:1.6.0-alpine go install -a  -ldflags "${BUILD}" -installsuffix cgo ./...

# Assemble common resource for ssl, timezones, and user.
mkdir -p docker-build-tmp/common/etc/ssl/certs
mkdir -p docker-build-tmp/common/usr/share
echo "nobody:x:65534:65534:Nobody:/:" > docker-build-tmp/common/etc/passwd
cp /etc/ssl/certs/ca-certificates.crt docker-build-tmp/common/etc/ssl/certs
# An alternative is to use $GOROOT/lib/time/zoneinfo.zip
rsync --archive /usr/share/zoneinfo docker-build-tmp/common/usr/share

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
