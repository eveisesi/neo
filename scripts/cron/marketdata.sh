#!/bin/sh

/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml up marketdata
/snap/bin/docker-compose -f /root/projects/neo/core/cron.yaml rm -s -f marketdata