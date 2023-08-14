---
title: "Templating"
---

{{< lead >}}
Generally speaking the templating uses [Golang `text/template`](https://pkg.go.dev/text/template) template syntax. All fields with templating enabled do support the full synax from the `text/template` package.
{{< /lead >}}

## Variables

There are certain variables available in the strings with templating enabled:

- `channel` - Channel the message was sent to, only available for regular messages not events
- `msg` - The message object, used in functions, should not be sent to chat
- `permitTimeout` - Value of `permit_timeout` in seconds
- `username` - The username of the message author


## Functions

Within templates following functions can be used:

- built-in functions in `text/template` engine
- functions from [sprig](https://masterminds.github.io/sprig/) function collection
- functions mentioned below

Examples below are using this syntax in the code block:

```
! Message matcher used for the input message
> Input message if used in the example
# Template used in the fields
< Output from the template
```

### `arg`

Takes the message sent to the channel, splits by space and returns the Nth element

Syntax: `arg <index>`

Example:

```
> !bsg @tester
# {{ arg 1 }} please refrain from BSG
< @tester please refrain from BSG
```

### `botHasBadge`

Checks whether bot has the given badge in the current channel

Syntax: `botHasBadge <badge>`

Example:

```
# {{ botHasBadge "moderator" }}
< true
```

### `channelCounter`

Wraps the counter name into a channel specific counter name including the channel name

Syntax: `channelCounter <counter name>`

Example:

```
# {{ channelCounter "test" }}
< 5
```

### `counterValue`

Returns the current value of the counter which identifier was supplied

Syntax: `counterValue <counter name>`

Example:

```
# {{ counterValue (list .channel "test" | join ":") }}
< 5
```

### `counterValueAdd`

Adds the given value (or 1 if no value) to the counter and returns its new value

Syntax: `counterValueAdd <counter name> [increase=1]`

Example:
```
# {{ counterValueAdd "myCounter" }} {{ counterValueAdd "myCounter" 5 }}
< 1 6
```

### `displayName`

Returns the display name the specified user set for themselves

Syntax: `displayName <username> [fallback]`

Example:

```
# {{ displayName "luziferus" }} - {{ displayName "notexistinguser" "foobar" }}
< Luziferus - foobar
```

### `doesFollow`

Returns whether `from` follows `to`

Syntax: `doesFollow <from> <to>`

Example:

```
# {{ doesFollow "tezrian" "luziferus" }}
< true
```

### `doesFollowLongerThan`

Returns whether `from` follows `to` for more than `duration`

Syntax: `doesFollowLongerThan <from> <to> <duration>`

Example:

```
# {{ doesFollowLongerThan "tezrian" "luziferus" "168h" }}
< true
```

### `fixUsername`

Ensures the username no longer contains the `@` or `#` prefix

Syntax: `fixUsername <username>`

Example:

```
# {{ fixUsername .channel }} - {{ fixUsername "@luziferus" }}
< luziferus - luziferus
```

### `formatDuration`

Returns a formated duration. Pass empty strings to leave out the specific duration part.

Syntax: `formatDuration <duration> <hours> <minutes> <seconds>`

Example:

```
# {{ formatDuration (streamUptime .channel) "hours" "minutes" "seconds" }} - {{ formatDuration (streamUptime .channel) "hours" "minutes" "" }}
< 5 hours, 33 minutes, 12 seconds - 5 hours, 33 minutes
```

### `followAge`

Looks up when `from` followed `to` and returns the duration between then and now

Syntax: `followAge <from> <to>`

Example:

```
# {{ followAge "tezrian" "luziferus" }}
< 15004h14m59.116620989s
```

### `followDate`

Looks up when `from` followed `to`

Syntax: `followDate <from> <to>`

Example:

```
# {{ followDate "tezrian" "luziferus" }}
< 2021-04-10 16:07:07 +0000 UTC
```

### `group`

Gets matching group specified by index from `match_message` regular expression, when `fallback` is defined, it is used when group has an empty match

Syntax: `group <idx> [fallback]`

Example:

```
! !command ([0-9]+) ([a-z]+) ([a-z]*)
> !command 12 test
# {{ group 2 "oops" }} - {{ group 3 "oops" }}
< test - oops
```

### `inList`

Tests whether a string is in a given list of strings (for conditional templates).

Syntax: `inList "search" "item1" "item2" [...]`

Example:

```
! !command (.*)
> !command foo
# {{ inList (group 1) "foo" "bar" }}
< true
```

### `jsonAPI`

Fetches remote URL and applies jq-like query to it returning the result as string. (Remote API needs to return status 200 within 5 seconds.)

Syntax: `jsonAPI "https://example.com/doc.json" ".data.exampleString" ["fallback"]`

Example:

```
! !mycmd
> !mycmd
# {{ jsonAPI "https://example.com/doc.json" ".data.exampleString" }}
< example string
```

### `lastPoll`

Gets the last (currently running or archived) poll for the given channel (the channel must have given extended permission for poll access!)

Syntax: `lastPoll <channel>`

Example:

```
# Last Poll: {{ (lastPoll .channel).Title }}
< Last Poll: Und wie siehts im Template aus?
```

See schema of returned object in [`pkg/twitch/polls.go#L13`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/polls.go#L13)

### `lastQuoteIndex`

Gets the last quote index in the quote database for the current channel

Syntax: `lastQuoteIndex`

Example:

```
# Last Quote: #{{ lastQuoteIndex }}
< Last Quote: #32
```

### `mention`

Strips username and converts into a mention

Syntax: `mention <username>`

Example:

```
# {{ mention "@user" }} {{ mention "user" }} {{ mention "#user" }}
< @user @user @user
```

### `pow`

Returns float from calculation: `float1 ** float2`

Syntax: `pow <float1> <float2>`

Example:

```
# {{ printf "%.0f" (pow 10 4) }}%
< 10000
```

### `randomString`

Randomly picks a string from a list of strings

Syntax: `randomString "a" [...]`

Example:

```
# {{ randomString "a" "b" "c" "d" }}
< a
```

### `recentGame`

Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.

Syntax: `recentGame <username> [fallback]`

Example:

```
# {{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}
< Metro Exodus - none
```


### `recentTitle`

Returns the last stream title of the specified user or the `fallback` if the title could not be fetched. If no fallback was supplied the message will fail and not be sent.

Syntax: `recentTitle <username> [fallback]`

Example:

```
# {{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}
< Die Oper haben wir überlebt, mal sehen was uns sonst noch alles töten möchte… - none
```

### `seededRandom`

Returns a float value stable for the given seed

Syntax: `seededRandom <string-seed>`

Example:

```
# Your int this hour: {{ printf "%.0f" (mul (seededRandom (list "int" .username (now | date "2006-01-02 15") | join ":")) 100) }}%
< Your int this hour: 17%
```

### `streamUptime`

Returns the duration the stream is online (causes an error if no current stream is found)

Syntax: `streamUptime <username>`

Example:

```
# {{ formatDuration (streamUptime "luziferus") "hours" "minutes" "" }}
< 3 hours, 56 minutes
```

### `subCount`

Returns the number of subscribers (accounts) currently subscribed to the given channel

Syntax: `subCount <channel>`

Example:

```
# {{ subCount "luziferus" }}
< 26
```

### `subPoints`

Returns the number of sub-points currently given through the T1 / T2 / T3 subscriptions to the given channel

Syntax: `subPoints <channel>`

Example:

```
# {{ subPoints "luziferus" }}
< 26
```

### `tag`

Takes the message sent to the channel, returns the value of the tag specified

Syntax: `tag <tagname>`

Example:

```
# {{ tag "login" }}
< luziferus
```

### `textAPI`

Fetches remote URL and returns the result as string. (Remote API needs to return status 200 within 5 seconds.)

Syntax: `textAPI "https://example.com/" ["fallback"]`

Example:

```
! !weather (.*)
> !weather Hamburg
# {{ textAPI (printf "https://api.scorpstuff.com/weather.php?units=metric&city=%s" (urlquery (group 1))) }}
< Weather for Hamburg, DE: Few clouds with a temperature of 22 C (71.6 F). [...]
```

### `variable`

Returns the variable value or default in case it is empty

Syntax: `variable <name> [default]`

Example:

```
# {{ variable "foo" "fallback" }} - {{ variable "unsetvar" "fallback" }}
< test - fallback
```

##  Upgrade from `v2.x` to `v3.x`

When adding [sprig](https://masterminds.github.io/sprig/) function collection some functions collided and needed replacement. You need to adapt your templates accordingly:

- Math functions (`add`, `div`, `mod`, `mul`, `multiply`, `sub`) were replaced with their sprig-equivalent and are now working with integers instead of floats. If you need them to continue to work with floats you need to use their [float-variants](https://masterminds.github.io/sprig/mathf.html).
- `now` does no longer format the current date as a string but return the current date. You need to replace this: `now "2006-01-02"` becomes `now | date "2006-01-02"`.
- `concat` is now used to concat arrays. To join strings you will need to modify your code: `concat ":" "string1" "string2"` becomes `lists "string1" "string2" | join ":"`.
- `toLower` / `toUpper` need to be replaced with their sprig equivalent `lower` and `upper`.
