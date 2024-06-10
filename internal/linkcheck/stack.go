package linkcheck

import "strings"

type (
	stack struct {
		visits []string
	}
)

func (s stack) Count(url string) (n int) {
	for _, v := range s.visits {
		if strings.EqualFold(v, url) {
			n++
		}
	}

	return n
}

func (s stack) Height() int {
	return len(s.visits)
}

func (s *stack) Visit(url string) {
	s.visits = append(s.visits, url)
}
