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
