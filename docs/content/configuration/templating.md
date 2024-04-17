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
< Output from the template (Rendered during docs generation)
* Output from the template (Static output, template not rendered)
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

### `b64urldec`

Decodes the input using base64 URL-encoding (like `b64dec` but using `URLEncoding` instead of `StdEncoding`)

Syntax: `b64urldec <input>`

Example:

```
# {{ b64urldec "bXlzdHJpbmc=" }}
< mystring
```

### `b64urlenc`

Encodes the input using base64 URL-encoding (like `b64enc` but using `URLEncoding` instead of `StdEncoding`)

Syntax: `b64urlenc <input>`

Example:

```
# {{ b64urlenc "mystring" }}
< bXlzdHJpbmc=
```

### `botHasBadge`

Checks whether bot has the given badge in the current channel

Syntax: `botHasBadge <badge>`

Example:

```
# {{ botHasBadge "moderator" }}
< false
```

### `channelCounter`

Wraps the counter name into a channel specific counter name including the channel name

Syntax: `channelCounter <counter name>`

Example:

```
# {{ channelCounter "test" }}
< #example:test
```

### `chatterHasBadge`

Checks whether chatter writing the current line has the given badge in the current channel

Syntax: `chatterHasBadge <badge>`

Example:

```
# {{ chatterHasBadge "moderator" }}
< true
```

### `counterRank`

Returns the rank of the given counter and the total number of counters in given counter prefix

Syntax: `counterRank <prefix> <name>`

Example:

```
# {{ $cr := counterRank (list .channel "test" "" | join ":") (list .channel "test" "foo" | join ":") }}{{ $cr.Rank }}/{{ $cr.Count }}
* 2/6
```

### `counterTopList`

Returns the top n counters for the given prefix as objects with Name and Value fields. Can be ordered by `name` / `value` / `first_seen` / `last_modified` ascending (`ASC`) or descending (`DESC`): i.e. `last_modified DESC` defaults to `value DESC`

Syntax: `counterTopList <prefix> <n> [orderBy]`

Example:

```
# {{ range (counterTopList (list .channel "test" "" | join ":") 3) }}{{ .Name }}: {{ .Value }} - {{ end }}
* #example:test:foo: 5 - #example:test:bar: 4 - 
```

### `counterValue`

Returns the current value of the counter which identifier was supplied

Syntax: `counterValue <counter name>`

Example:

```
# {{ counterValue (list .channel "test" | join ":") }}
* 5
```

### `counterValueAdd`

Adds the given value (or 1 if no value) to the counter and returns its new value

Syntax: `counterValueAdd <counter name> [increase=1]`

Example:

```
# {{ counterValueAdd "myCounter" }} {{ counterValueAdd "myCounter" 5 }}
* 1 6
```

### `displayName`

Returns the display name the specified user set for themselves

Syntax: `displayName <username> [fallback]`

Example:

```
# {{ displayName "luziferus" }} - {{ displayName "notexistinguser" "foobar" }}
* Luziferus - foobar
```

### `doesFollow`

Returns whether `from` follows `to` (the bot must be moderator of `to` to read this)

Syntax: `doesFollow <from> <to>`

Example:

```
# {{ doesFollow "tezrian" "luziferus" }}
* true
```

### `doesFollowLongerThan`

Returns whether `from` follows `to` for more than `duration` (the bot must be moderator of `to` to read this)

Syntax: `doesFollowLongerThan <from> <to> <duration>`

Example:

```
# {{ doesFollowLongerThan "tezrian" "luziferus" "168h" }}
* true
```

### `fixUsername`

Ensures the username no longer contains the `@` or `#` prefix

Syntax: `fixUsername <username>`

Example:

```
# {{ fixUsername .channel }} - {{ fixUsername "@luziferus" }}
< example - luziferus
```

### `followAge`

Looks up when `from` followed `to` and returns the duration between then and now (the bot must be moderator of `to` to read this)

Syntax: `followAge <from> <to>`

Example:

```
# {{ followAge "tezrian" "luziferus" }}
* 15004h14m59.116620989s
```

### `followDate`

Looks up when `from` followed `to` (the bot must be moderator of `to` to read this)

Syntax: `followDate <from> <to>`

Example:

```
# {{ followDate "tezrian" "luziferus" }}
* 2021-04-10 16:07:07 +0000 UTC
```

### `formatDuration`

Returns a formated duration. Pass empty strings to leave out the specific duration part.

Syntax: `formatDuration <duration> <hours> <minutes> <seconds>`

Example:

```
# {{ formatDuration .testDuration "hours" "minutes" "seconds" }} - {{ formatDuration .testDuration "hours" "minutes" "" }}
< 5 hours, 33 minutes, 12 seconds - 5 hours, 33 minutes
```

### `formatHumanDateDiff`

Formats a DateInterval object according to the format (%Y, %M, %D, %H, %I, %S for years, months, days, hours, minutes, seconds - Lowercase letters without leading zeros)

Syntax: `formatHumanDateDiff <format> <obj>`

Example:

```
# {{ humanDateDiff (mustToDate "2006-01-02 -0700" "2024-05-05 +0200") (mustToDate "2006-01-02 -0700" "2023-01-09 +0100") | formatHumanDateDiff "%Y years, %M months, %D days" }}
< 01 years, 03 months, 25 days
```

### `group`

Gets matching group specified by index from `match_message` regular expression, when `fallback` is defined, it is used when group has an empty match

Syntax: `group <idx> [fallback]`

Example:

```
! !command ([0-9]+) ([a-z]+) ?([a-z]*)
> !command 12 test
# {{ group 2 "oops" }} - {{ group 3 "oops" }}
< test - oops
```

### `humanDateDiff`

Returns a DateInterval object describing the time difference between a and b in a "human" way of counting the time (2023-02-05 -> 2024-03-05 = 1 Year, 1 Month)

Syntax: `humanDateDiff <a> <b>`

Example:

```
# {{ humanDateDiff (mustToDate "2006-01-02 -0700" "2024-05-05 +0200") (mustToDate "2006-01-02 -0700" "2023-01-09 +0100") }}
< {1 3 25 23 0 0}
```

### `idForUsername`

Returns the user-id for the given username

Syntax: `idForUsername <username>`

Example:

```
# {{ idForUsername "twitch" }}
* 12826
```

### `inList`

Tests whether a string is in a given list of strings (for conditional templates).

Syntax: `inList <search> <...string>`

Example:

```
! !command (.*)
> !command foo
# {{ inList (group 1) "foo" "bar" }}
< true
```

### `jsonAPI`

Fetches remote URL and applies jq-like query to it returning the result as string. (Remote API needs to return status 200 within 5 seconds.)

Syntax: `jsonAPI <url> <jq-like path> [fallback]`

Example:

```
# {{ jsonAPI "https://api.github.com/repos/Luzifer/twitch-bot" ".owner.login" }}
* Luzifer
```

### `lastPoll`

Gets the last (currently running or archived) poll for the given channel (the channel must have given extended permission for poll access!)

Syntax: `lastPoll <channel>`

Example:

```
# Last Poll: {{ (lastPoll .channel).Title }}
* Last Poll: Und wie siehts im Template aus?
```

See schema of returned object in [`pkg/twitch/polls.go#L13`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/polls.go#L13)

### `lastQuoteIndex`

Gets the last quote index in the quote database for the current channel

Syntax: `lastQuoteIndex`

Example:

```
# Last Quote: #{{ lastQuoteIndex }}
* Last Quote: #32
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
# {{ printf "%.0f" (pow 10 4) }}
< 10000
```

### `profileImage`

Gets the URL of the given users profile image

Syntax: `profileImage <username>`

Example:

```
# {{ profileImage .username }}
* https://static-cdn.jtvnw.net/jtv_user_pictures/[...].png
```

### `randomString`

Randomly picks a string from a list of strings

Syntax: `randomString <string> [...string]`

Example:

```
# {{ randomString "a" "b" "c" "d" }}
* a
```

### `recentGame`

Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.

Syntax: `recentGame <username> [fallback]`

Example:

```
# {{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}
* Metro Exodus - none
```

### `recentTitle`

Returns the last stream title of the specified user or the `fallback` if the title could not be fetched. If no fallback was supplied the message will fail and not be sent.

Syntax: `recentTitle <username> [fallback]`

Example:

```
# {{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}
* Die Oper haben wir überlebt, mal sehen was uns sonst noch alles töten möchte… - none
```

### `scheduleSegments`

Returns the next n segments in the channels schedule. If n is not given, returns all known segments.

Syntax: `scheduleSegments <channel> [n]`

Example:

```
# {{ $seg := scheduleSegments "luziferus" 1 | first }}Next Stream: {{ $seg.Title }} @	{{ dateInZone "2006-01-02 15:04" $seg.StartTime "Europe/Berlin" }}
* Next Stream: Little Nightmares @ 2023-11-05 18:00
```

### `seededRandom`

Returns a float value stable for the given seed

Syntax: `seededRandom <string-seed>`

Example:

```
# Your int this hour: {{ printf "%.0f" (mulf (seededRandom (list "int" .username (now | date "2006-01-02 15") | join ":")) 100) }}%
< Your int this hour: 72%
```

### `spotifyCurrentPlaying`

Retrieves the current playing track for the given channel

Syntax: `spotifyCurrentPlaying <channel>`

Example:

```
! ^!spotify
> !spotify
# {{ spotifyCurrentPlaying .channel }}
* Beast in Black - Die By The Blade
```

### `spotifyLink`

Retrieves the link for the playing track for the given channel

Syntax: `spotifyLink <channel>`

Example:

```
! ^!spotifylink
> !spotifylink
# {{ spotifyLink .channel }}
* https://open.spotify.com/track/3HCzXf0lNpekSqsGBcGrCd
```

### `streamUptime`

Returns the duration the stream is online (causes an error if no current stream is found)

Syntax: `streamUptime <username>`

Example:

```
# {{ formatDuration (streamUptime "luziferus") "hours" "minutes" "" }}
* 3 hours, 56 minutes
```

### `subCount`

Returns the number of subscribers (accounts) currently subscribed to the given channel

Syntax: `subCount <channel>`

Example:

```
# {{ subCount "luziferus" }}
* 26
```

### `subPoints`

Returns the number of sub-points currently given through the T1 / T2 / T3 subscriptions to the given channel

Syntax: `subPoints <channel>`

Example:

```
# {{ subPoints "luziferus" }}
* 26
```

### `tag`

Takes the message sent to the channel, returns the value of the tag specified

Syntax: `tag <tagname>`

Example:

```
# {{ tag "display-name" }}
< ExampleUser
```

### `textAPI`

Fetches remote URL and returns the result as string. (Remote API needs to return status 200 within 5 seconds.)

Syntax: `textAPI <url> [fallback]`

Example:

```
! !weather (.*)
> !weather Hamburg
# {{ textAPI (printf "https://api.scorpstuff.com/weather.php?units=metric&city=%s" (urlquery (group 1))) }}
* Weather for Hamburg, DE: Few clouds with a temperature of 22 C (71.6 F). [...]
```

### `userExists`

Checks whether the given user exists

Syntax: `userExists <username>`

Example:

```
# {{ userExists "luziferus" }}
* true
```

### `usernameForID`

Returns the current login name of an user-id

Syntax: `usernameForID <user-id>`

Example:

```
# {{ usernameForID "12826" }}
* twitch
```

### `variable`

Returns the variable value or default in case it is empty

Syntax: `variable <name> [default]`

Example:

```
# {{ variable "foo" "fallback" }} - {{ variable "unsetvar" "fallback" }}
* test - fallback
```
