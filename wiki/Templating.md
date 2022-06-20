## Templating

Generally speaking the templating uses [Golang `text/template`](https://pkg.go.dev/text/template) template syntax. All fields with templating enabled do support the full synax from the `text/template` package.

### Variables

There are certain variables available in the strings with templating enabled:

- `channel` - Channel the message was sent to, only available for regular messages not events
- `msg` - The message object, used in functions, should not be sent to chat
- `permitTimeout` - Value of `permit_timeout` in seconds
- `username` - The username of the message author


### Functions

Additionally to the built-in functions there are extra functions available in the templates:

Examples below are using this syntax in the code block:

```
! Message matcher used for the input message
> Input message if used in the example
# Template used in the fields
< Output from the template
```

#### `add`

Returns float from calculation: `float1 + float2`

Syntax: `add <float1> <float2>`

Example:

```
# {{ printf "%.0f" (add 1 2) }}%
< 3
```

#### `arg`

Takes the message sent to the channel, splits by space and returns the Nth element

Syntax: `arg <index>`

Example:

```
> !bsg @tester
# {{ arg 1 }} please refrain from BSG
< @tester please refrain from BSG
```

#### `botHasBadge`

Checks whether bot has the given badge in the current channel

Syntax: `botHasBadge <badge>`

Example:

```
# {{ botHasBadge "moderator" }}
< true
```

#### `channelCounter`

Wraps the counter name into a channel specific counter name including the channel name

Syntax: `channelCounter <counter name>`

Example:

```
# {{ channelCounter "test" }}
< 5
```

#### `concat`

Join the given string parts with delimiter

Syntax: `concat <delimiter> <...parts>`

Example:

```
# {{ concat ":" "test" .username }}
< test:luziferus
```

#### `counterValue`

Returns the current value of the counter which identifier was supplied

Syntax: `counterValue <counter name>`

Example:

```
# {{ counterValue (concat ":" .channel "test") }}
< 5
```

#### `displayName`

Returns the display name the specified user set for themselves

Syntax: `displayName <username> [fallback]`

Example:

```
# {{ displayName "luziferus" }} - {{ displayName "notexistinguser" "foobar" }}
< Luziferus - foobar
```

#### `div`

Returns float from calculation: `float1 / float2`

Syntax: `div <float1> <float2>`

Example:

```
# {{ printf "%.0f" (div 27 9) }}%
< 3
```

#### `fixUsername`

Ensures the username no longer contains the `@` or `#` prefix

Syntax: `fixUsername <username>`

Example:

```
# {{ fixUsername .channel }} - {{ fixUsername "@luziferus" }}
< luziferus - luziferus
```

#### `formatDuration`

Returns a formated duration. Pass empty strings to leave out the specific duration part.

Syntax: `formatDuration <duration> <hours> <minutes> <seconds>`

Example:

```
# {{ formatDuration (streamUptime .channel) "hours" "minutes" "seconds" }} - {{ formatDuration (streamUptime .channel) "hours" "minutes" "" }}
< 5 hours, 33 minutes, 12 seconds - 5 hours, 33 minutes
```

#### `followDate`

Looks up when `from` followed `to`

Syntax: `followDate <from> <to>`

Example:

```
# {{ followDate "tezrian" "luziferus" }}
< 2021-04-10 16:07:07 +0000 UTC
```

#### `group`

Gets matching group specified by index from `match_message` regular expression, when `fallback` is defined, it is used when group has an empty match

Syntax: `group <idx> [fallback]`

Example:

```
! !command ([0-9]+) ([a-z]+) ([a-z]*)
> !command 12 test
# {{ group 2 "oops" }} - {{ group 3 "oops" }}
< test - oops
```

#### `inList`

Tests whether a string is in a given list of strings (for conditional templates).

Syntax: `inList "search" "item1" "item2" [...]`

Example:

```
! !command (.*)
> !command foo
# {{ inList (group 1) "foo" "bar" }}
< true
```

#### `lastQuoteIndex`

Gets the last quote index in the quote database for the current channel

Syntax: `lastQuoteIndex`

Example:

```
# Last Quote: #{{ lastQuoteIndex }}
< Last Quote: #32
```

#### `mul` (deprecated: `multiply`)

Returns float from calculation: `float1 * float2`

Syntax: `mul <float1> <float2>`

Example:

```
# {{ printf "%.0f" (mul 100 (seededRandom "test")) }}%
< 35%
```

#### `pow`

Returns float from calculation: `float1 ** float2`

Syntax: `pow <float1> <float2>`

Example:

```
# {{ printf "%.0f" (pow 10 4) }}%
< 10000
```

#### `recentGame`

Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.

Syntax: `recentGame <username> [fallback]`

Example:

```
# {{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}
< Metro Exodus - none
```

#### `streamUptime`

Returns the duration the stream is online (causes an error if no current stream is found)

Syntax: `streamUptime <username>`

Example:

```
# {{ formatDuration (streamUptime "luziferus") "hours" "minutes" "" }}
< 3 hours, 56 minutes
```

#### `seededRandom`

Returns a float value stable for the given seed

Syntax: `seededRandom <string-seed>`

Example:

```
# Your int this hour: {{ printf "%.0f" (multiply (seededRandom (concat ":" "int" .username (now "2006-01-02 15"))) 100) }}%
< Your int this hour: 17%
```

#### `sub`

Returns float from calculation: `float1 - float2`

Syntax: `sub <float1> <float2>`

Example:

```
# {{ printf "%.0f" (sub 10 4) }}%
< 6
```

#### `tag`

Takes the message sent to the channel, returns the value of the tag specified

Syntax: `tag <tagname>`

Example:

```
# {{ tag "login" }}
< luziferus
```

#### `toLower` / `toUpper`

Converts the given string to lower-case / upper-case

Syntax: `toLower <string>` / `toUpper <string>`

Example:

```
# {{ toLower "Test" }} - {{ toUpper "Test" }}
< test - TEST
```

#### `variable`

Returns the variable value or default in case it is empty

Syntax: `variable <name> [default]`

Example:

```
# {{ variable "foo" "fallback" }} - {{ variable "unsetvar" "fallback" }}
< test - fallback
```

