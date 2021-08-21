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
