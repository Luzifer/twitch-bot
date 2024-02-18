![](https://badges.fyi/github/license/Luzifer/twitch-bot)
![](https://badges.fyi/github/downloads/Luzifer/twitch-bot)
![](https://badges.fyi/github/latest-release/Luzifer/twitch-bot)

# Luzifer / twitch-bot

Twitch-Bot is intended as an alternative to having a bot managed by Streamlabs or Streamelements and therefore having more control over it, the availability and how it works.

At the moment it is a work-in-progress and does not yet implment all features it shall in the future.

## Configuration

Please refer to the [Documentation](https://luzifer.github.io/twitch-bot/) how to setup and configure the bot.

```console
# twitch-bot --help
Usage of twitch-bot:
      --base-url string                  External URL of the config-editor interface (used to generate auth-urls)
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
  actor-docs                                        Generate markdown documentation for available actors
  api-token <token-name> <scope> [...scope]         Generate an api-token to be entered into the config
  copy-database <target storage-type> <target DSN>  Copies database contents to a new storage DSN i.e. for migrating to a new DBMS
  reset-secrets                                     Remove encrypted data to reset encryption passphrase
  tpl-docs                                          Generate markdown documentation for available template functions
  validate-config                                   Try to load configuration file and report errors if any
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
