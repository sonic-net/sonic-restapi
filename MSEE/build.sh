
set -e

# copy debian packages from sonic-buildimage
bash copy.sh

# build a container with build utilities
docker build -t rest-api-build-image --rm -f Dockerfile.build .

# create a directory for target .deb files
mkdir -p /tmp/target

# create .deb packages inside of the build container
docker run --rm=true -v /tmp/target:/src -v $(pwd):/src/rest              -w /src/rest rest-api-build-image dpkg-buildpackage -us -uc
docker run --rm=true -v /tmp/target:/src -v $(pwd)/arp_responder:/src/arp -w /src/arp  rest-api-build-image dpkg-buildpackage -us -uc

# copy created packages to out debs directoy
cp /tmp/target/*.deb debs

# remove cruft
sudo rm -fr /tmp/target

# build rest-api test image
docker build -t rest-api-image --rm --squash .

# save rest-api-image into a file
docker save rest-api-image | gzip > rest-api-image.gz

## Production docker containers

# build rest-api prod common image
docker build -t rest-api-common --rm --squash -f Dockerfile.common.prod .

# build rest-api prod bm image
docker build -t rest-api-prod-bm --rm --squash -f Dockerfile.bm.prod .

# build rest-api prod msee image
docker build -t rest-api-prod-msee --squash --rm -f Dockerfile.msee.prod .

# save rest-api-prod-bm into a file
docker save rest-api-prod-bm | gzip > rest-api-prod-bm.gz

# save rest-api-prod-msee into a file
docker save rest-api-prod-msee | gzip > rest-api-prod-msee.gz
