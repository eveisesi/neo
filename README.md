## New Eden Obituary (NEO)

Welcome. NEO is a golang port of [ZKillboard](https://zkillboard)[[Github](https://github.com/zkillboard/zkillboard)]. I started this project as a response to a port by Squizz on Reddit. Me and him talk quite often on Tweetfleet Slack and he has been requesting a competitor to ZKillboard for quite a while. After talking with another friend who agreed to assist me with the startup funding i would need for the server, we were able to start development on NEO. I build NEO so that anybody who has a box laying around, can run it. All you need is Redis, MySQL, Docker, and about 6GB of RAM. I go into more detail about installing it below.

### Frontend

[NEO's UI](https://neo.eveisesi.space)[[Github](https://github.com/eveisesi/neo-ui)] is written in VueJS and is the frontend that talks to the NEO API. I build this website up independantly of the backend so that I could potentially swap it out at a later date and so that you can do whatever you want. Run it yourself or write your own.

### Getting up and Running with NEO

There are couple of host dependencies that NEO requires outside of docker. I initially attempted to run these inside of Docker, but do to the initial load requirements of the project (importing millions of killmails), I was constantly losing connection to Redis and MySQL when they were inside Docker Containers. For this reason, NEO's Redis and MySQL Dependencies currently run on the Host itself.

### Environment Variables

NEO has network_mode set to host to facilitate talking to the Redis and MySQL instance on the host.
So DBHOST and REDISHOST below are set to `127.0.0.1:<PORT>`

```
DBUSER=<string>
DBPASS=<string>
DBHOST=<ip address:port>
DBNAME=<string>
DBREADTIMEOUT=<int|defaults to 30>
DBWRITETIMEOUT=<int|defaults to 30>

Env=<string|enum:[production, development]>

LOGLEVEL=<string|ref logrus for valid values, but info, err, and debug are the only values used in this porject>

ESIHOST=esi.evetech.net
ESIUAGENT=<string>

REDISADDR=<ip address:port>

# Zkillboard User Agent
ZUAGENT=<string>

SERVER_PORT=<int|don't set this below 5000>

# https://developers.eveonline.com/
SSO_CLIENT_ID=<string>
SSO_CLIENT_SECRET=<string>
SSO_CALLBACK=<string>
SSO_AUTHORIZATION_URL=https://login.eveonline.com/v2/oauth/authorize
SSO_TOKEN_URL=https://login.eveonline.com/v2/oauth/token
SSO_JWKS_URL=https://login.eveonline.com/oauth/jwks

# https://api.slack.com/messaging/webhooks
SLACK_NOTIFIER_ENABLED=<bool>
SLACK_NOTIFIER_URL=<string>
# Killmails valued at or above this number will be posted to the provided webhook URL
SLACK_NOTIFIER_THRESHOLD=<int|multiplied by 1K>
# Base URL for Action buttons in webhook message (Format: https://example.com)
SLACK_ACTION_BASE_URL=<string>

# Backup for Processed Killmail pre DB write. Completely optional
# https://www.digitalocean.com/products/spaces/
SPACES_ENABLED=<bool>
SPACES_BUCKET=<string>
SPACES_ENDPOINT=<string>
SPACES_REGION=<string>
SPACES_KEY=<string>
SPACES_SECRET=<string>
```

docker.env

```
export DOCKER_VERSION="0.17.8" # Github Tag
export PROCESS_LIMIT=10 # Num of GoRoutines to use
export PROCESS_SLEEP=100 # Milliseconds Routine should sleep before returning when done processing
export RECAL_QUEUE_LIMIT=10000
export RECAL_QUEUE_MINIMUM=5000
export RECAL_NUM_WORKERS=50 # Num of GoRoutines to use
# Loops from MAX -> MIN in descending order
export HISTORY_MAX=20191031 # Date Year Month Day
export HISTORY_MIN=20150101 # Same as above
# Holds on each date rather than just continuing to loop and put everything on the queue. Use in low memory situations
export HISTORY_DATEHOLD=true
# if datehold is true, when the queue gets below this number, the loop will continue to the next date
export HISTORY_THRESHOLD=200
export BACKUP_PROCESS_LIMIT=7 # Num of GoRoutines to use
export BACKUP_PROCESS_SLEEP=100 # Millisecond Routine sleeps before returning when do processing

```

---

### Nginx

For security reasons I will not expose the nginx configurations for NEO, but I will assist by saying that most of the services are setup on a subdomain and then use nginx proxies to the containers themselves.

---

### Redis

```
~/projects/neo# redis-server -v
Redis server v=6.0.0

```

#### Plugins

Outside of NEO's normal uses for Redis, we also store an index inside of Redis for the search functionality and then leverage the RedisSearch plugin to facilitate those searches. You will need to have this plugin install else search will not work and I believe the cron would panic

[RedisSearch](https://github.com/RediSearch/RediSearch)

---

### MySQL

Typically MySQL installation

`Server version: 5.7`

#### Migrations

NEO has a migration system that can executed to facilitate initial configuration of the schema

`docker-compose -f docker-compose-init.yaml up && docker-compose -f docker-compose-int.yaml down`

---
