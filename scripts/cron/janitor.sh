#!/bin/sh

/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml up janitor
/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml rm -s -f janitor