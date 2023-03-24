package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type (
	cliRegistry struct {
		cmds map[string]cliRegistryEntry
		sync.Mutex
	}

	cliRegistryEntry struct {
		Description string
		Name        string
		Params      []string
		Run         func([]string) error
	}
)

var (
	cli           = newCLIRegistry()
	errHelpCalled = errors.New("help called")
)

func newCLIRegistry() *cliRegistry {
	return &cliRegistry{
		cmds: make(map[string]cliRegistryEntry),
	}
}

func (c *cliRegistry) Add(e cliRegistryEntry) {
	c.Lock()
	defer c.Unlock()

	c.cmds[e.Name] = e
}

func (c *cliRegistry) Call(args []string) error {
	c.Lock()
	defer c.Unlock()

	cmdEntry := c.cmds[args[0]]
	if cmdEntry.Name != args[0] {
		c.help()
		return errHelpCalled
	}

	return cmdEntry.Run(args)
}

func (c *cliRegistry) help() {
	// Called from Call, does not need lock

	var (
		maxCmdLen int
		cmds      []cliRegistryEntry
	)

	for name := range c.cmds {
		entry := c.cmds[name]
		if l := len(entry.CommandDisplay()); l > maxCmdLen {
			maxCmdLen = l
		}
		cmds = append(cmds, entry)
	}

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

	tpl := fmt.Sprintf("  %%-%ds  %%s\n", maxCmdLen)
	fmt.Fprintln(os.Stdout, "Supported sub-commands are:")
	for _, cmd := range cmds {
		fmt.Fprintf(os.Stdout, tpl, cmd.CommandDisplay(), cmd.Description)
	}
}

func (c cliRegistryEntry) CommandDisplay() string {
	return strings.Join(append([]string{c.Name}, c.Params...), " ")
}
