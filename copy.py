#!/usr/local/bin/python3

import json
import os
import shutil
import urllib.parse
from subprocess import Popen, PIPE

URL = "https://sonic-build.azurewebsites.net/api/sonic/artifacts"
CONF_FILE = "dependencies.conf"

def main():
    conf_file_path = os.path.join(CONF_FILE)
    conf_t = open(conf_file_path).read()
    config = json.loads(conf_t)

    if not os.path.exists("./debs/"):
        os.mkdir("./debs")

    swsscommon_debs_dir = os.environ.get("SWSSCOMMON_DEBS_DIR")
    for dep in config['dependencies']:
        if swsscommon_debs_dir and dep.startswith(("libswsscommon_", "libswsscommon-dev_")):
            filename = f"{dep}_{config['arch']}.deb"
            shutil.copyfile(os.path.join(swsscommon_debs_dir, filename), os.path.join("debs", filename))
            continue

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
