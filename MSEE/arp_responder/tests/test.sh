#!/bin/bash

if [[ $(id -u) -ne 0 ]];
then
  echo Please run as root
  exit
fi

docker images | grep '^autoresponder_test' > /dev/null
if [[ $? -ne 0 ]];
then
    docker build --no-cache \
             --rm=true \
             -t autoresponder_test \
             build
fi

function add_if()
{
  num=$1
  pid=$2
  extif=eif$num
  intif=iif$num
  ip link add $extif type veth peer name $intif
  ip link set netns $pid dev $intif
  ip link set $extif up
  nsenter -t $pid -n ip link set $intif up
}

# Start docker container
docker run -d -it -p9091:9091/tcp --name arpresponder_test --cap-add NET_ADMIN autoresponder_test > /dev/null

# Insert interfaces inside of the container
docker_id=arpresponder_test
pid=$(docker inspect --format '{{.State.Pid}}' $docker_id)
add_if 0 $pid
add_if 1 $pid

echo

# Compile
cd ..
debian/rules clean
dpkg-buildpackage -us -uc
docker cp ../arpresponder-bm_1.0.0_amd64.deb arpresponder_test:/tmp
docker cp ../arpresponder-msee_1.0.0_amd64.deb arpresponder_test:/tmp
docker exec -ti arpresponder_test dpkg -i /tmp/arpresponder-bm_1.0.0_amd64.deb
docker exec -ti arpresponder_test dpkg -i /tmp/arpresponder-msee_1.0.0_amd64.deb
cd tests

# Test MSEE
thrift --gen py ../arp_responder.thrift 
mv gen-py/arp_responder ptf_tests/
rm -fr gen-py
docker exec -ti arpresponder_test supervisorctl start arpresponder_msee > /dev/null
sudo ptf --disable-ipv6 --test-dir ptf_tests msee_tests -i 0-0@eif0 -i 0-1@eif1
docker exec -ti arpresponder_test supervisorctl stop arpresponder_msee > /dev/null

echo

# Test Baremetal
docker exec -ti arpresponder_test bash -c "echo 'iif0' > /tmp/arpresponder.conf"
docker exec -ti arpresponder_test bash -c "echo 'iif1' >> /tmp/arpresponder.conf"
docker exec -ti arpresponder_test supervisorctl start arpresponder_bm > /dev/null
sudo ptf --disable-ipv6 --test-dir ptf_tests bm_tests -i 0-0@eif0 -i 0-1@eif1
docker exec -ti arpresponder_test supervisorctl stop arpresponder_bm > /dev/null

# Clean up
rm -fr ptf_tests/arp_responder
rm -f ptf.log
rm -f ptf.pcap
docker stop arpresponder_test > /dev/null
docker rm arpresponder_test > /dev/null
