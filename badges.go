package main

type badgeCollection map[string]*int

func (b badgeCollection) Add(badge string, level int) {
	b[badge] = &level
}

func (b badgeCollection) Get(badge string) int {
	l, ok := b[badge]
	if !ok {
		return 0
	}

	return *l
}

func (b badgeCollection) Has(badge string) bool {
	return b[badge] != nil
}
