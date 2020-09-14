#!/bin/sh

/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml up tqstatus
/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml rm -s -f tqstatus