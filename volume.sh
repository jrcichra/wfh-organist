#!/usr/bin/env bash
set -x
set -o pipefail

VOLUME=$1
DEVICE=$(pactl list sink-inputs | grep -B 17 'application.name = "ALSA plug-in \[rx\]"' | grep 'Sink Input #' | cut -d'#' -f2)
if [ $? -ne 0 ]; then
    echo "No rx in pulseaudio found"
    exit 1
fi
pactl set-sink-input-volume ${DEVICE} ${VOLUME}%