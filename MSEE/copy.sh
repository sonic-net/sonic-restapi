set -e

SDIR=/data/sonic_build/sonic-buildimage/target/debs
cp $SDIR/libhiredis0.13_0.13.3-2_amd64.deb debs
cp $SDIR/libhiredis-dev_0.13.3-2_amd64.deb debs
cp $SDIR/libnl-3-200_3.2.27-2_amd64.deb debs
cp $SDIR/libnl-genl-3-200_3.2.27-2_amd64.deb debs
cp $SDIR/libnl-route-3-200_3.2.27-2_amd64.deb debs
cp $SDIR/libswsscommon_1.0.0_amd64.deb debs
cp $SDIR/libswsscommon-dev_1.0.0_amd64.deb debs
cp $SDIR/libthrift-0.9.3_0.9.3-2_amd64.deb debs
