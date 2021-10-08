#!/usr/bin/env python

import os, signal
from subprocess import check_output
import time

time.sleep(10*60)

pid = int(check_output(["pidof", "go-server-server.test"]))
os.kill(pid, signal.SIGQUIT)
