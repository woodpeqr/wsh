package flags

import (
	"V-Woodpecker-V/wsh/warg/internal/log"
	"fmt"
	"strings"
)

func ParseArgs(args []string) error {
	pArgs := preprocessArgs(args)

	var curValueFlag *WFlag
	curFlagContext := flagRegistry

	for _, arg := range pArgs {
		var f *WFlag
		if strings.HasPrefix(arg, "-") {
			for f == nil {
				f = matchFlag(curFlagContext, arg)
			}
		}
		if f == nil {
			if curValueFlag == nil || (strings.HasPrefix(arg, "-") && !strings.Contains(arg, " ")) {
				log.Error(fmt.Sprintf("unknown argument: %s", arg))
				return fmt.Errorf("unknown argument: %s", arg)
			}
			curValueFlag.setValue(arg)
		} else {
			f.setValue(true)
			if f.ValueRequired || f.NonEmptyValueRequired {
				curValueFlag = f
			}
		}
	}
	return nil
}

func preprocessArgs(args []string) []string {
	processedArgs := []string{}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			processedArgs = append(processedArgs, strings.Trim(arg, " "))
		} else {
			for _, char := range []rune(arg)[1:] {
				if char == ' ' {
					continue
				}
				processedArgs = append(processedArgs, fmt.Sprintf("-%c", char))
			}
		}
	}
	return processedArgs
}

func matchFlag(flags []*WFlag, arg string) *WFlag {
	for _, wFlag := range flags {
		a := strings.TrimLeft(arg, "-")
		if (strings.HasPrefix(arg, "--") && a == wFlag.Long) ||
			(strings.HasPrefix(arg, "-") && a == wFlag.Short) {
			return wFlag
		}
	}
	return nil
}
