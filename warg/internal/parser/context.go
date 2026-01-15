package parser

import (
	"V-Woodpecker-V/wsh/warg/flags"
)

// Context represents a parsing context with its flag definitions
type Context struct {
	Flags  map[string]*flags.FlagDefinition
	Parent *Context
}

// NewContext creates a new parsing context from flag definitions
func NewContext(defs []flags.FlagDefinition, parent *Context) *Context {
	ctx := &Context{
		Flags:  make(map[string]*flags.FlagDefinition),
		Parent: parent,
	}

	for i := range defs {
		def := &defs[i]
		for _, name := range def.Names {
			ctx.Flags[name] = def
		}
	}

	return ctx
}

// Lookup searches for a flag in the current context and parent contexts
func (c *Context) Lookup(name string) *flags.FlagDefinition {
	if def, ok := c.Flags[name]; ok {
		return def
	}
	if c.Parent != nil {
		return c.Parent.Lookup(name)
	}
	return nil
}
