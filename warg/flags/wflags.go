package flags

import (
	"fmt"
	"reflect"
	"strconv"
)

var flagRegistry []*WFlag

type WFlag struct {
	Short                 string
	Long                  string
	Help                  string
	Parent                *WFlag
	Children              []*WFlag
	ValueRequired         bool
	NonEmptyValueRequired bool
	ptr                   any
}

func AddFlag(flag *WFlag) {
	v := reflect.ValueOf(flag.ptr)
	if v.Kind() != reflect.Pointer {
		panic("flag.ptr must be a pointer")
	}

	flagRegistry = append(flagRegistry, flag)
}

func AddFlags(flags []*WFlag) {
	for _, flag := range flags {
		AddFlag(flag)
	}
}

func DebugPrintFlags() {
	for _, f := range flagRegistry {
		fmt.Printf("-%s --%s - '%s'\n", f.Short, f.Long, reflect.ValueOf(f.ptr).Elem().Interface())
	}
}

func (w *WFlag) setValue(val any) error {
	p := reflect.ValueOf(w.ptr).Elem()
	v := reflect.ValueOf(val)
	switch p.Kind() {
	case reflect.String:
		if p.Kind() != v.Kind() {
			return fmt.Errorf("invalid value type for string flag: %s", v.Kind().String())
		}
		p.Set(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Kind() == reflect.String {
			i, err := strconv.ParseInt(v.String(), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value for int flag: %s", v.String())
			}
			p.SetInt(i)
		} else if v.Kind() != p.Kind() {
			return fmt.Errorf("invalid value type for int flag: %s", v.Kind().String())
		} else {
			p.Set(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Kind() == reflect.String {
			u, err := strconv.ParseUint(v.String(), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value for uint flag: %s", v.String())
			}
			p.SetUint(u)
		} else if v.Kind() != p.Kind() {
			return fmt.Errorf("invalid value type for uint flag: %s", v.Kind().String())
		} else {
			p.Set(v)
		}
	case reflect.Bool:
		if v.Kind() == reflect.String {
			b, err := strconv.ParseBool(v.String())
			if err != nil {
				return fmt.Errorf("invalid value for bool flag: %s", v.String())
			}
			p.SetBool(b)
		} else if v.Kind() != reflect.Bool {
			return fmt.Errorf("invalid value type for bool flag: %s", v.Kind().String())
		} else {
			p.Set(v)
		}
	case reflect.Slice:
		if v.Kind() != reflect.Slice {
			return fmt.Errorf("invalid value type for slice flag: %s", v.Kind().String())
		}
		if p.Type().Elem() != v.Type().Elem() {
			return fmt.Errorf("invalid slice element type for flag: %s", v.Type().Elem().String())
		}
		p.Set(reflect.AppendSlice(p, v))
	default:
		return fmt.Errorf("unsupported flag type: %s", p.Kind().String())
	}
	return nil
}
