#!/usr/bin/env bash
# Run this from crontab to keep a wireguard port open
set -x
INTERFACE=$(route -n  | grep '^0.0.0.0' | head -1 | awk '{print $8}')
IP=$(ip addr | grep ${INTERFACE} | grep inet | awk '{print $2}' | cut -d"/" -f1)
WIREGUARD_PORT=51820
HOURS=3 # your cron should be faster than this
upnpc -a ${IP} ${WIREGUARD_PORT} ${WIREGUARD_PORT} udp $((HOURS*60*60))
