#!/bin/sh
set -e

# Xvfb provides the display headed Chrome needs
Xvfb :99 -screen 0 1920x1080x24 -nolisten tcp >/dev/null 2>&1 &
for _ in $(seq 1 50); do
	[ -e /tmp/.X11-unix/X99 ] && break
	sleep 0.1
done

exec "$@"
