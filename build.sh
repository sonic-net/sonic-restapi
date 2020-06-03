
set -e

# copy debian packages from sonic-buildimage
bash copy.sh

# build a container with build utilities
docker build -t rest-api-build-image --rm -f Dockerfile.build .

# create a directory for target .deb files
mkdir -p /tmp/target

# create .deb packages inside of the build container
docker run --rm=true -v /tmp/target:/src -v $(pwd):/src/rest -w /src/rest rest-api-build-image dpkg-buildpackage -us -uc
#docker run --rm=true -v /tmp/target:/src -v $(pwd)/arp_responder:/src/arp -w /src/arp  rest-api-build-image dpkg-buildpackage -us -uc

# copy created packages to out debs directoy
cp /tmp/target/*.deb debs

# remove cruft
sudo rm -fr /tmp/target

# build rest-api test image
docker build -t rest-api-image-test_local --rm -f Dockerfile.test .
