---
title: Service Setup
weight: 3
---

> [!TIP]
> In order to have your bot started automatically on system start we should set it up as a service.

## General preparation

> [!INFO]
> In order not to put confidential information into the configuration file we want to create a configuration file to hold these secrets. Also we will create a folder to put stored data and overlay files into. You can use other folders for those files, just remember to adjust the paths in all places. For this page I will assume the binary is placed in `/usr/local/bin/twitch-bot_linux_amd64`.

```console
# First create a folder to hold our secret file(s)
$ mkdir /etc/twitch-bot

# Second create a folder to hold the data
$ mkdir -p /var/lib/twitch-bot/overlays

# Create the secrets file and secure access to it
$ touch /etc/twitch-bot/environment
$ chown root:root /etc/twitch-bot/environment
$ chmod 0600 /etc/twitch-bot/environment

# Edit the file to hold the secrets (use your favourite editor instead of nano)
$ nano /etc/twitch-bot/environment
```

Lets put this into the file we've just edited (replace the brackets, don't leave them inside the file!):

```env
BASE_URL=http://localhost:3000/

CONFIG=/var/lib/twitch-bot/config.yaml
LOG_LEVEL=info
OVERLAYS_DIR=/var/lib/twitch-bot/overlays

STORAGE_CONN_STRING=/var/lib/twitch-bot/storage.db
STORAGE_CONN_TYPE=sqlite
STORAGE_ENCRYPTION_PASS=[put a random secure password here]

TWITCH_CLIENT=[put the client-id from the preparations step here]
TWITCH_CLIENT_SECRET=[put the client-secret from the preparations step here]
```

If you want to use a database server like MariaDB or PostgreSQL adjust the `STORAGE_CONN_*` variables like described in the project [README](https://github.com/Luzifer/twitch-bot/blob/master/README.md#database-connection-strings).

## Option 1: Using a downloaded or compiled Binary

We will create a systemd service to start the binary using the environment variables file we've created in the preparation step. This file will be placed into `/etc/systemd/system/twitch-bot.service`:

```service
[Unit]
Description=Twitch-Bot Service
After=network-online.target
Requires=network-online.target

[Service]
EnvironmentFile=/etc/twitch-bot/environment
ExecStart=/usr/local/bin/twitch-bot_linux_amd64
Restart=Always
RestartSecs=5

[Install]
WantedBy=multi-user.target
```

To enable and start the service which makes it automatically start on every server boot execute these commands:

```console
$ systemctl daemon-reload
$ systemctl enable --now twitch-bot
```

After the first start a configuration file has been created at `/var/lib/twitch-bot/config.yaml`. You want to change the port in the `http_listen` line in this file:

```yaml
# IP/Port to start the web-interface on. Format: IP:Port
# The default is 127.0.0.1:0 - Listen on localhost with random port
http_listen: "127.0.0.1:3000"
```

After changing the port restart the service once:

```console
$ systemctl restart twitch-bot
```

To update the bot first stop the bot, then replace the binary and start the bot again:

```console
$ systemctl stop twitch-bot
$ mv [path to new binary] /usr/local/bin/twitch-bot_linux_amd64
$ systemctl start twitch-bot
```

## Option 2: Using a Docker image

We will create a systemd service to start the binary using the environment variables file we've created in the preparation step. This file will be placed into `/etc/systemd/system/twitch-bot.service`:

```service
[Unit]
Description=Twitch-Bot Service
After=network-online.target
Requires=network-online.target

[Service]
EnvironmentFile=/etc/twitch-bot/environment
ExecStartPre=-/usr/bin/docker rm -f %n
ExecStartPre=/usr/bin/docker pull luzifer/twitch-bot:stable
ExecStart=/usr/bin/docker run --rm --name %n \
            --env-file /etc/twitch-bot/environment \
            -v /var/lib/twitch-bot:/var/lib/twitch-bot \
            -p 127.0.0.1:3000:3000 \
            luzifer/twitch-bot:stable
Restart=Always
RestartSecs=5

[Install]
WantedBy=multi-user.target
```

To enable and start the service which makes it automatically start on every server boot execute these commands:

```console
$ systemctl daemon-reload
$ systemctl enable --now twitch-bot
```

After the first start a configuration file has been created at `/var/lib/twitch-bot/config.yaml`. You want to change IP and port in the `http_listen` line in this file:

```yaml
# IP/Port to start the web-interface on. Format: IP:Port
# The default is 127.0.0.1:0 - Listen on localhost with random port
http_listen: "0.0.0.0:3000"
```

After changing the port restart the service once:

```console
$ systemctl restart twitch-bot
```

To update the bot just restart the service:

```console
$ systemctl restart twitch-bot
```

## Debugging the Service

In both options you created a service which is running as a system user. You can do the same things for both options to see what the bot is doing:

```console
$ journalctl -fu twitch-bot
```

This will stream the logs of the bots into your terminal and you can read what the bot is doing.

In case the bot behaves unexpectedly you can increase the `LOG_LEVEL` from `info` to `debug` within the `/etc/twitch-bot/environment` file and restart the bot to get more verbose logs.
