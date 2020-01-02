#!/bin/bash

/sbin/brctl addbr Ethernet12
/sbin/ifconfig Ethernet12 up
/sbin/brctl addbr Ethernet16
/sbin/ifconfig Ethernet16 up
/sbin/brctl addbr Ethernet34
/sbin/ifconfig Ethernet34 up
