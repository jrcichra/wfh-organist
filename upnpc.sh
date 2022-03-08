#!/usr/bin/env bash
# Run this from crontab to keep a wireguard port open
set -x
IP=$(ip addr | grep eth0 | grep inet | awk '{print $2}' | cut -d"/" -f1)
WIREGUARD_PORT=51820
HOURS=3 # your cron should be faster than this
upnpc -a ${IP} ${WIREGUARD_PORT} ${WIREGUARD_PORT} udp $((HOURS*60))
