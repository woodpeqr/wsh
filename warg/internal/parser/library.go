package parser

import (
	"fmt"
	
	"V-Woodpecker-V/wsh/warg/flags"
)

// ParseWithSetters parses args with the given definitions and applies setter functions
// This function is designed to be called from the flags package to avoid import cycles
func ParseWithSetters(defs []flags.FlagDefinition, args []string, setters map[string]func(string) error) error {
	// Use the existing parser
	p := NewParser(defs)
	parseResult, err := p.Parse(args)
	if err != nil {
		return err
	}

	// Apply setters to parsed values
	var walkErr error
	parseResult.Walk(func(fv *FlagValue) {
		if walkErr != nil {
			return
		}

		// Find setter for this flag
		var setterFn func(string) error
		for _, name := range fv.Definition.Names {
			if fn, ok := setters[name]; ok {
				setterFn = fn
				break
			}
		}

		if setterFn == nil {
			return
		}

		// Apply the setter
		if fv.Definition.Switch {
			// For switch flags, pass "true" if present
			if fv.Present {
				if err := setterFn("true"); err != nil {
					walkErr = fmt.Errorf("error setting flag %v: %w", fv.Definition.Names[0], err)
				}
			}
		} else {
			// For value flags, pass the value
			if fv.Present {
				if err := setterFn(fv.Value); err != nil {
					walkErr = fmt.Errorf("error setting flag %v: %w", fv.Definition.Names[0], err)
				}
			}
		}
	})

	return walkErr
}
