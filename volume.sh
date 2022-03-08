#!/usr/bin/env bash
set -x

VOLUME=$1
DEVICE=$(pactl list sink-inputs | grep -B 17 'application.name = "ALSA plug-in \[rx\]"' | grep 'Sink Input #' | cut -d'#' -f2)

pactl set-sink-input-volume ${DEVICE} ${VOLUME}%