package linkcheck

import (
	"fmt"
)

const (
	chromeMajor = 128
	webkitMajor = 537
	webkitMinor = 36
)

// generateUserAgent resembles the Chrome user agent generation as
// closely as possible in order to blend into the crowd of browsers
//
// https://github.com/chromium/chromium/blob/58e23d958ee8d2bb4b085c843a18eb28b9da17da/content/common/user_agent.cc
func generateUserAgentHeaders() map[string]string {
	return map[string]string{
		// New UA hints method
		"Sec-CH-UA": fmt.Sprintf(
			`"Chromium";v="%[1]d", "Not;A=Brand";v="24", "Google Chrome";v="%[1]d"`,
			chromeMajor,
		),

		// Not a mobile browser
		"Sec-CH-UA-Mobile": "?0",

		// We're always Windows
		"Sec-CH-UA-Platform": "Windows",

		// "old" user-agent
		"User-Agent": fmt.Sprintf(
			"Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) %s Safari/537.36",
			"Windows NT 10.0; Win64; x64",               // We're always Windows 10 / 11 on x64
			fmt.Sprintf("Chrome/%d.0.0.0", chromeMajor), // UA-Reduction enabled
		),
	}
}
