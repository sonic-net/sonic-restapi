#!/usr/local/bin/python3

import json
import os
import urllib.parse
from subprocess import Popen, PIPE

URL = "https://sonic-build.azurewebsites.net/api/sonic/artifacts"
CONF_FILE = "dependencies.conf"
DOWNLOAD_CMD = "wget -O debs/libhiredis0.14_0.14.0-3~bpo9+1_amd64.deb  'https://sonic-build.azurewebsites.net/api/sonic/artifacts?branchName=master&platform=vs&target=target%2Fdebs%2Fbuster%2Flibhiredis0.14_0.14.0-3~bpo9%2B1_amd64.deb'"

def main():
    conf_file_path = os.path.join(CONF_FILE)
    conf_t = open(conf_file_path).read()
    config = json.loads(conf_t)

    if not os.path.exists("./debs/"):
        os.mkdir("./debs")

    for dep in config['dependencies']:   
        params = {
            'branchName': config['branch'],
            'platform': config['platform'],
            'target': "target/debs/"+config['debian']+"/"+dep+"_"+config['arch']+".deb"
            }
        url = URL+"?"+urllib.parse.urlencode(params)
        download_cmd = "wget -O debs/"+dep+"_"+config['arch']+".deb "+url
        print(download_cmd)
        process = Popen(download_cmd.split(), stdout=PIPE, stderr=PIPE)
        process.communicate()

if __name__ == "__main__":
    main()
