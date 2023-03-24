[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/twitch-bot)](https://goreportcard.com/report/github.com/Luzifer/twitch-bot)
![](https://badges.fyi/github/license/Luzifer/twitch-bot)
![](https://badges.fyi/github/downloads/Luzifer/twitch-bot)
![](https://badges.fyi/github/latest-release/Luzifer/twitch-bot)
![](https://knut.in/project-status/twitch-bot)

# Luzifer / twitch-bot

Twitch-Bot is intended as an alternative to having a bot managed by Streamlabs or Streamelements and therefore having more control over it, the availability and how it works.

At the moment it is a work-in-progress and does not yet implment all features it shall in the future.

## Configuration

Please see the [Wiki](https://github.com/Luzifer/twitch-bot/wiki) for documentation of the configuration file.

```console
# twitch-bot --help
Usage of twitch-bot:
      --base-url string                  External URL of the config-editor interface (set to enable EventSub support)
      --command-timeout duration         Timeout for command execution (default 30s)
  -c, --config string                    Location of configuration file (default "./config.yaml")
      --log-level string                 Log level (debug, info, warn, error, fatal) (default "info")
      --plugin-dir string                Where to find and load plugins (default "/usr/lib/twitch-bot")
      --rate-limit duration              How often to send a message (default: 20/30s=1500ms, if your bot is mod everywhere: 100/30s=300ms, different for known/verified bots) (default 1.5s)
      --sentry-dsn string                Sentry / GlitchTip DSN for error reporting
      --storage-conn-string string       Connection string for the database (default "./storage.db")
      --storage-conn-type string         One of: mysql, postgres, sqlite (default "sqlite")
      --storage-encryption-pass string   Passphrase to encrypt secrets inside storage (defaults to twitch-client:twitch-client-secret)
      --twitch-client string             Client ID to act as
      --twitch-client-secret string      Secret for the Client ID
      --twitch-token string              OAuth token valid for client (fallback if no token was set in interface)
  -v, --validate-config                  Loads the config, logs any errors and quits with status 0 on success
      --version                          Prints current version and exits
      --wait-for-selfcheck duration      Maximum time to wait for the self-check to respond when behind load-balancers (default 1m0s)

# twitch-bot help
Supported sub-commands are:
  actor-docs                                 Generate markdown documentation for available actors
  api-token <token-name> <scope> [...scope]  Generate an api-token to be entered into the config
  migrate-v2 <old-file>                      Migrate old (*.json.gz) storage file into new database
  reset-secrets                              Remove encrypted data to reset encryption passphrase
  validate-config                            Try to load configuration file and report errors if any
```

### Database Connection Strings

Currently these databases are supported and need their corresponding connection strings:

#### MySQL

```
[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]

Recommended parameters:
  ?charset=utf8mb4&parseTime=True&loc=Local
```

- Create your database as follows:  
  ```sql
  CREATE DATABASE twbot_tezrian DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci;
  ```
- Start your bot:  
  ```console
  # twitch-bot \
      --storage-conn-type mysql \
      --storage-conn-string 'tezrian:mypass@tcp(mariadb:3306)/twbot_tezrian?charset=utf8mb4&parseTime=True&loc=Local' \
      ...
  ```

See [driver documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for more details on parameters.

#### Postgres

```
host=localhost port=5432 dbname=mydb connect_timeout=10
```

See [Postgres documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS) for more details in paramters.

#### SQLite

```
storage.db
```

Just pass the filename you want to use.

- Start your bot:  
  ```console
  # twitch-bot \
      --storage-conn-type sqlite \
      --storage-conn-string 'storage.db' \
      ...
  ```


## Upgrade from `v2.x` to `v3.x`

With the release of `v3.0.0` the bot changed a lot introducing a new storage format. As that storage backend is not compatible with the `v2.x` storage you need to migrate it manually before starting a `v3.x` bot version the first time.

**Before starting the migration make sure to fully stop the bot!**

This section assumes you were starting your `v2.x` bot the following way:

```console
# twitch-bot \
  --storage-file storage.json.gz
  --twitch-client <clientid> \
  --twitch-client-secret <secret>
```

To execute the migration we need to provide the same `storage-encryption-pass` or `twitch-client` / `twitch-client-secret` combination if no `storage-encryption-pass` was used.

```console
# twitch-bot \
  --storage-conn-type <database type> \
  --storage-conn-string <database connection string> \
  --twitch-client <clientid> \
  --twitch-client-secret <secret> \
  migrate-v2 storage.json.gz
WARN[0000] No storage encryption passphrase was set, falling back to client-id:client-secret
WARN[0000] Module registered unhandled query-param type  module=status type=integer
WARN[0000] Overlays dir not specified, no dir or non existent  dir=
INFO[0000] Starting migration...                         module=variables
INFO[0000] Starting migration...                         module=mod_punish
INFO[0000] Starting migration...                         module=mod_overlays
INFO[0000] Starting migration...                         module=mod_quotedb
INFO[0000] Starting migration...                         module=core
INFO[0000] Starting migration...                         module=counter
INFO[0000] Starting migration...                         module=permissions
INFO[0000] Starting migration...                         module=timers
INFO[0000] v2 storage file was migrated
```

If you see the `v2 storage file was migrated` message the contents of your old storage file were migrated to the new database. The old file is not modified in this step.

Afterwards your need to adjust the start parameters of the bot:

```console
# twitch-bot \
  --storage-conn-type <database type> \
  --storage-conn-string <database connection string> \
  --twitch-client <clientid> \
  --twitch-client-secret <secret> \
```
