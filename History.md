# 3.18.2 / 2023-10-08

  * Bugfixes
    * [core] Fix: New followers endpoint requires user-token

# 3.18.1 / 2023-10-05

  * Bugfixes
    * [core] Fix: Replace deprecated follow API

# 3.18.0 / 2023-09-21

  * New Features
    * [core] Add channel specific module configuration interface
    * [templating] Add `idForUsername` function
    * [templating] Add `usernameForID` function

  * Improvements
    * [core] Add `user:manage:whispers` extended scope
    * [core] Update go-irc to v4.0.0

  * Bugfixes
    * [ci] Update dependencies
    * [raffle] Insert newly created raffles with `NULL` reminder time

  * Documentation
    * [docs] Add raffle documentation
    * [docs] Add raffle module as feature to start page
    * [docs] Fix broken preparations image

  * Deprecations
    * [core] Mark twitch-token flag / envvar deprecated
    * [core] Remove v2 migration

# 3.17.0 / 2023-08-25

  * New Features
    * [templating] Add `b64urldec` and `b64urlenc` functions

  * Improvements
    * [docs] Add auto-generated template documentation (#50)

  * Bugfixes
    * [ci] Remove flaky test

> [!WARNING]  
> This marks the last release to contain code to migrate from v2.x to v3.x releases. If you are migration from an old v2 instance at a later point in time you need to migrate to this version before continuing your journey to the latest v3 release.

# 3.16.0 / 2023-08-22

  * New Features
    * [clip] Add `clip` actor
    * [messagehook] Add actor for Discord / Slack hook posts
    * [overlays] Add `sounds` overlay as default
    * [templating] Add `profileImage` function

  * Improvements
    * [docs] Move documentation from Wiki to docs-site (#49)
    * [docs] Add Apache2 config sample (thanks to @Breitling1992)
    * [docs] Add "VIP of the Month" example rule (thanks to @Breitling1992)

  * Bugfixes
    * [core] Fix: Clean usernames when querying user information
    * [editor] Add `shoutout_created` to frontend-known events

# 3.15.0 / 2023-08-04

  * New Features
    * [core] Add support for `hype_chat` event

  * Improvements
    * [eventsub] Switch to `channel.update/2`
    * [linkdetector] Add new option to enable heuristic scan
    * [twitchclient] Reduce retries and errors when banning banned user

# 3.14.2 / 2023-07-21

  * Bugfixes
    * [ban] Fix Chatcommand matching

# 3.14.1 / 2023-07-16

  * Bugfixes
    * [raffle] Fix index initialization in MySQL v8

# 3.14.0 / 2023-07-16

  * New Features
    * Implement Raffle module (#47)
    * [template] Add `textAPI` function

  * Improvements
    * [ci] Update nodejs version for builds
    * [eventsub] Replace `IsMature` tag in channel updates

  * Bugfixes
    * [wiki] Fix example broken since v3.x

# 3.13.0 / 2023-06-25

  * New Features
    * [counter] Add `counterValueAdd` template function

  * Improvements
    * [core] Add cleanup for expired timers
    * [core] Clean IPs from socket errors

  * Bugfixes
    * [core] Fix missing timer configuration for permits

# 3.12.0 / 2023-06-07

  * New Features
    * [respond] Expose API route to send messages directly to chat
    * [template] Add `mention` function

  * Improvements
    * [eventsub] Add `status` field to `poll_end` event

# 3.11.0 / 2023-05-27

  * New Features
    * [eventsub] Add `poll_begin`, `poll_end`, `poll_progress` events
    * [template] Add `lastPoll` function

  * Improvements
    * [core] Reduce variance of Sentry errors containing IPs
    * [eventsub] Add debug logging for subscribed topics

# 3.10.0 / 2023-05-21

  * New Features
    * [eventsub] Switch to Websocket transport (#46)

  * Improvements
    * [core] Adjust logging for channel meta updates to match other events
    * [core] Allow case insensitive category matches
    * [editor] Remove character limit for AutoMessage template

# 3.9.0 / 2023-05-11

  * New Features
    * [template] Add `subCount`, `subPoints` template functions

  * Bugfixes
    * [wiki] Remove deprecated `concat` examples

# 3.8.0 / 2023-04-14

  * New Features
    * [linkprotect] Add Link-, Clip-Detector and Link-Protection actor (#42)

  * Improvements
    * [core] Add connection tuning for MySQL databases
    * [core] Remove "host" related functionality
    * [editor] Add validation for template fields

  * Bugfixes
    * [core] Fix: Message matcher matched also event content
    * [editor] Fix badge key-repetition for duplicated actions

# 3.7.0 / 2023-03-31

  * New Features
    * [commercial] Add `commercial` actor
    * [eventsub] Add `shoutout_created` event

  * Improvements
    * [core] Add validation and reset of encrypted values
    * [eventsub] Switch to v2 follows topic

  * Bugfixes
    * [core] Ensure channel has correct format in access service
    * [core] Fix: Allow start when no tokens are available
    * [core] Fix type warnings for Swagger documentation
    * [eventsub] Fix wrong channel in `shoutout_received` event

# 3.6.0 / 2023-03-06

  * New Features
    * [eventmod] Add `eventmod` actor
    * [eventsub] Add `shoutout_received` event

  * Improvements
    * [script] Add rule ID to error

  * Bugfixes
    * [editor] Fix number-of-lines mode causing type-error

# 3.5.1 / 2023-02-08

  * Bugfixes
    * [core] Fix: List all configured channel permissions

# 3.5.0 / 2023-02-08

  * New Features
    * [shield] Add shield mode actor
    * [stopexec] Add `stopexec` actor
    * [template] Add `recentTitle` template function

  * Improvements
    * [core] Rewrite bot token storage logic
    * [editor] Add new `moderator:read:followers` scope and pin follow subscription version
    * [editor] Notify frontend to reload data after token change

  * Bugfixes
    * [editor] Ensure updating bot token does not drop scopes
    * [editor] Fix Node package vulnerabilities
    * [editor] Fix non-optional booleans causing rules to be non-saveable
    * [editor] Fix: When `match_message` is cleared, remove it completely

# 3.4.0 / 2023-01-27

  * New Features
    * [shoutout] Implement actor and slash-command for shoutout API

  * Improvements
    * [editor] Add notification in case bot is missing default-scopes

# 3.3.0 / 2023-01-07

  * Bugfixes
    * [core] Fix: Remote-update cron broken as of missing field

  * New Features
    * [log] Add `log`-actor
    * [template] Add `doesFollow` and `doesFollowLongerThan` functions
    * [templating] Add `followAge` function

  * Improvements
    * [customevent] Add scheduled events to API handler

# 3.2.1 / 2022-12-24

  * Bugfixes
    * [twitch] Fix: Pagination fetching broken for eventsub subscriptions

# 3.2.0 / 2022-12-22

  * New Features
    * Add fine-grained permission control for extended channel permissions (#35)
    * [twitch] Implement `AddChannelVIP`, `RemoveChannelVIP`
    * [vip/unvip] Implement actors and chat commands

  * Improvements
    * [core] Add content-type detection for remote rule subscriptions
    * [core] Add retries for eventsub-self-check
    * [core] Add validation for rule UUIDs to be unique
    * [core] Allow plugins to evaluate whether permissions are available

# 3.1.0 / 2022-11-24

  * New Features
    * [core] Add Sentry / GlitchTip error reporting

# 3.0.0 / 2022-11-02

**⚠ Breaking Changes:**
  - Backend storage format has been switched from JSON-file to database. Migrations must be run before use of `v3.x` version. See [README](https://github.com/Luzifer/twitch-bot#upgrade-from-v2x-to-v3x) for instructions.
  - Some template function have been migrated to a new function collection. See [migration section of Templating documentation](https://github.com/Luzifer/twitch-bot/wiki/Templating#upgrade-from-v2x-to-v3x) for required changes.

**Changelog:**

  * New Features
    * [core] Add config validation command
    * [core] Add rule-subscription feature
    * [core] Add `outbound_raid` event
    * [customevent] Add scheduled custom events
    * [templating] Add `jsonAPI` template function

  * Improvements
    * [core] Move storage to database (#30, #32) ⚠
    * [core] Allow to pass ID to channel modification
    * [core] Extend API and replace deprecated chat commands (#34)
    * [editor] Add all template functions to highlighter
    * [overlays] Add `hide` option to debug overlay
    * [templating] Add sprig functions, replace some built-ins ⚠

  * Bugfixes
    * [core] Fix: Allow 5s for rule updates

# 2.7.1 / 2022-09-06

Bugfix release, repeating `v2.7.0` changelog as of broken release.

  * New Features
    * [template] Add `randomString` template function

  * Improvements
    * [core] Make number of subscribed months available for subgift
    * [security] Add mitigation for slowloris DoS attack vector

  * Bugfixes
    * [msgformatter] Fix: Trim leading / trailing spaces
    * [ci / lint] Fix missing CI tooling, fix linter errors

# 2.7.0 / 2022-09-03

  * New Features
    * [template] Add `randomString` template function

  * Improvements
    * [core] Make number of subscribed months available for subgift

  * Bugfixes
    * [msgformatter] Fix: Trim leading / trailing spaces

# 2.6.0 / 2022-07-15

  * New Features
    * [editor] [#18] Add editor for `disable_on_match_messages`
    * [template] Add `inList` function
    * [template] Add "mod" function for modulo in templating

  * Improvements
    * [core] Expose user\_id in events
    * [editor] Add explanatory hint for exceptions

  * Bugfixes
    * [editor] Fix: Token badges had no spacing

# 2.5.0 / 2022-06-06

  * Improvements
    * [core] Add multi\_month parameter parsing for subs

# 2.4.0 / 2022-05-07

  * Improvements
    * [editor] [#23] Add confirmation for delete buttons
    * [editor] [#25] Allow searching in / sort rules

  * Bugfixes
    * [core] Fix: Notify event handlers before rules to prevent delays
    * [editor] [#28] Fix: Allow saving with empty optional duration
    * [editor] Fix: Remove asymmetric margin from buttons
    * [modchannel] [#26] Fix: Modify channel module not working for editor-bots (#27)

# 2.3.0 / 2022-04-22

  * New Features
    * [core] Add more mathematical functions for templating
    * [customevent] Add API module and actor to create custom events
    * [filesay] Add FileSay actor to "paste" files with commands
    * [msgformat] Add module to retrieve filled template through API
    * [overlays] Add overlays server with some example templates and library

  * Improvements
    * [core] Add `delete` event for deleted chat messages
    * [core] Add `origin_id` to subgift / submysterygift events
    * [core] Add support for `annoumcement` event type
    * [core] add `total_gifted` field for gifts, use numeric values for some fields
    * [core] Provide message in `announcement`, `bits` and `resub` events
    * [counter] Add template support for counter step
    * [counter] Remove stored counter value on zero value
    * [editor] Add bot version to frontend
    * [editor] Improve location of permission warning
    * [timeout] [#15] Allow timeout reason to be set

  * Bugfixes
    * [ban] Fix: Add missing API docs
    * [core] Delete refresh token only for HTTP errors, not on connection issues
    * [core] Fix: Accept 1s cooldown, fix user and channel cooldowns
    * [core] Fix: EventSub messages had misformatted channel
    * [core] Fix: Handle unauthorized error for app-access-tokens
    * [core] Fix: Raid viewercount should be numeric, not string
    * [core] Re-check token validity more often than on expiry
    * [editor] [#19] Validate durations when checking for invalid rules
    * [editor] [#20] Fix: Strip query parameters from redirect uri
    * [editor] Fix node package vulnerability / update dependencies
    * [editor] Fix: Upgrade contains a header send, error must not send headers
    * [status] Fix: Add missing API docs

# 2.2.0 / 2022-01-16

  * [ci] Make installed go binaries available during build
  * [core] Add deprecated but still used V5 ChannelEditor scope
  * [core] Add EventSub subscription prefetching
  * [core] Add "follow" event using EventSub
  * [core] Add handling for channel point rewards
  * [core] Do not retry POST requests automatically
  * [core] Fix: Event data was not available in rule templates
  * [core] Implement dynamic token update and broadcaster permissions (#13)
  * [core] Improve EventSub API request design
  * [docs] Update README
  * [editor] Display disconnected status instead of error
  * [editor] Fix follow-redirects vulnerability (CVE-2022-0155)
  * [editor] Prevent adding invalid usernames as channel / editor

# 2.1.0 / 2021-12-24

  * [automessage] Add disable switch
  * [ban] Add HTTP API route for banning users
  * [core] Add status / health check API
  * [core] Fix: send-message function passed to plugin was nil
  * [core] Fix: Strip newlines from message templates
  * [core] log bits from chat message
  * [editor] Fix: Removing cooldown resulted in save error
  * [editor] Rework to use esbuild / Vue component files (#12)

# 2.0.0 / 2021-12-03

  * [ban] Enable templating for ban reason
  * [core] Add `giftpaidupgrade` event
  * [core] **BREAKING:** Allow actors to set fields those after them (#11)
  * [core] Fix: Set channel for incoming host through jtv message
  * [core] Handle host announce messages from jtv user
  * [lint] Properly format inputs
  * [templating] Add `multiply` and `seededRandom` template functions

# 1.6.0 / 2021-11-11

  * [core] Add `ban`, `clearchat` and `timeout` events
  * [core] Add EventSub support for Twitch-Events (#10)
  * [core] Add moderator badge to broadcasters
  * [core] Prevent logging every PING message

# 1.5.0 / 2021-11-04

  * [nuke] Add new moderation module

# 1.4.0 / 2021-10-25

  * [core] Allow the bot to track config editor changes through Git
  * [core] Implement write authorization for APIs (#9)
  * [editor] Cleanup config by removing invalid / zero attributes
  * [openapi] Allow multiple mime-types on single route
  * [plugins] Move missing plugin-dir warning to debug level
  * [quotedb] Add simple page to list quotes

# 1.3.0 / 2021-10-22

  * [core] Add "bits" event
  * [core] Add `streamUptime` / `formatDuration` template functions
  * [core] Add submysterygift event, add more event data to events
  * [core] Add username fields to events
  * [core] Remove unused subscribed\_months field from subgifts
  * [openapi] Allow subdir serving
  * [quotedb] Add new actor
  * [respond] Fix: Broken condition for fallback message
  * [respond] Fix: Empty string fallback should not count as fallback
  * [respond] Log message template errors even when fallback is set

# 1.2.0 / 2021-10-08

  * [core] Log submysterygift
  * [automessage] Move spammy message to trace-level
  * [core] Improve logs for USERNOTICE events
  * [editor] Add description to "Add Action" form group
  * Add "punish", "reset-punish" actors and storage manager (#8)

# 1.1.0 / 2021-10-01

  * [templating] Add `botHasBadge` function
  * [editor] Mark fully disabled rules in list

# 1.0.0 / 2021-09-22

  * Breaking: Add configuration interface and switch to more generic config format (#7)

# 0.18.0 / 2021-09-17

  * [script] Allow to skip cooldown on script error
  * [modchannel] Add modchannel core module
  * [core] Break actions execution when one action fails
  * [core] Transform broadcaster name into ID
  * [core] Add category search and channel update
  * [core] Expose GetIDForUsername function
  * [core] Expose TwitchClient to plugins
  * [core] Add fallback support to group template matches
  * [respond] Support sending message to different channel
  * [core] Reduce cache time for stream info
  * [core] Add Twitch events
  * [core] Add registration for raw-message-handlers

# 0.17.0 / 2021-08-28

  * Create API for counter and setvariable modules
  * Provide HTTP server and registration function
  * Provide central cron service to plugins

# 0.16.0 / 2021-08-21

  * Update dependencies and bring plugin example to work with master
  * Lint: Ignore gocritic for fatal program exit not running unlock
  * Move to Go1.17 mod-file, update dependencies
  * Disable CGO for default container
  * Allow plugins to register template functions
  * Add plugin support to allow extending of functionality (#6)
  * Add support to disable cooldown through the action module
  * Add method to send messages from within the bot without trigger
  * Add validation mode for config

# 0.15.0 / 2021-06-30

  * Wiki: Add example for generic chat-addable commands
  * Add support for dynamic variables
  * Lint: Update linter list, disable gomnd for some lines
  * Move timers to storage to persist them
  * Fix: Set channel for more events

# 0.14.0 / 2021-06-17

  * Fix: JSON is not able to decode `2s` but `2` which is ns instead of s
  * Add concat template function
  * Disable auto-messages in non-observed channels
  * Automatically leave channel when removed from config

# 0.13.0 / 2021-06-13

  * Use more flexible Actor format to allow addition of new actors (#5)
  * Add user- and channel-based cooldowns (#4)
  * Fix: ID generation handling different automessages as same
  * Fix: Do not try to log functions
  * Fix: Do not access automessage attributes without lock

# 0.12.0 / 2021-06-05

  * Add "respond as reply" functionality

# 0.11.0 / 2021-06-02

  * Add retries to Twitch API calls

# 0.10.0 / 2021-05-27

  * Add Whisper / RawMessage actions
  * Add `whisper` event

# 0.9.0 / 2021-05-26

  * Add `part` event
  * Allow to disable automessages with templates
  * Add global variables to be used in templates
  * Add Disable and DisableOnTemplate attributes for rules
  * Drop HCL support (causes too much effort for too little benefit)

# 0.8.0 / 2021-05-24

  * Fix: Display fallback when no category is set in `recentGame`
  * Add displayName template function
  * Replace non-reliable fsevents library with simple check
  * Add HCL config format support

# 0.7.0 / 2021-05-13

  * Lint: Disable requirement for crypto/rand for time randomizer
  * Add delay-action

# 0.6.0 / 2021-05-12

  * Add sub events, document available event types

# 0.5.0 / 2021-05-11

  * Fix: Unlock auto-messages to prevent dead-locks
  * Log amount of loaded rules on (re)load
  * Support templating in automessages

# 0.4.1 / 2021-05-06

  * Include tzdata into Docker image to allow TZ env setting
  * Update README for new flags

# 0.4.0 / 2021-04-22

  * Introduce general send limit to prevent global-timeouts

# 0.3.0 / 2021-04-21

  * Extract template functions into registry
  * Lint: Reduce complexity of loadConfig function
  * Add raw-log functionality
  * Add a delay while joining channels

# 0.2.0 / 2021-04-04

  * Add instructions for token generation
  * Add GH page to generate token

# 0.1.0 / 2021-04-03

  * Initial release
