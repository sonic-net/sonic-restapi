set -e

wget -O debs/libhiredis0.14_0.14.0-3~bpo9+1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target%2Fdebs%2Fbuster%2Flibhiredis0.14_0.14.0-3~bpo9%2B1_amd64.deb'
wget -O debs/libhiredis-dev_0.14.0-3~bpo9+1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target%2Fdebs%2Fbuster%2Flibhiredis-dev_0.14.0-3~bpo9%2B1_amd64.deb'
wget -O debs/libnl-3-200_3.5.0-1_amd64.deb 'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libnl-3-200_3.5.0-1_amd64.deb'
wget -O debs/libnl-3-dev_3.5.0-1_amd64.deb 'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libnl-3-dev_3.5.0-1_amd64.deb'
wget -O debs/libnl-genl-3-200_3.5.0-1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libnl-genl-3-200_3.5.0-1_amd64.deb'
wget -O debs/libnl-route-3-200_3.5.0-1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libnl-route-3-200_3.5.0-1_amd64.deb'
wget -O debs/libnl-nf-3-200_3.5.0-1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libnl-nf-3-200_3.5.0-1_amd64.deb'
wget -O debs/libthrift-0.11.0_0.11.0-4_amd64.deb 'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target/debs/buster/libthrift-0.11.0_0.11.0-4_amd64.deb'
wget -O debs/libswsscommon_1.0.0_amd64.deb 'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target%2Fdebs%2Fbuster%2Flibswsscommon_1.0.0_amd64.deb'
wget -O debs/libswsscommon-dev_1.0.0_amd64.deb 'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target%2Fdebs%2Fbuster%2Flibswsscommon-dev_1.0.0_amd64.deb'
