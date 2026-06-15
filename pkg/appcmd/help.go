package appcmd

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// fieldDesc describes a single configurable field extracted from a config struct.
type fieldDesc struct {
	Name    string // display name (field/json tag, or lowercased Go field name)
	Flag    string // CLI flag name without "--" (cli tag)
	EnvVar  string // full environment variable name (env tag + accumulated envPrefix)
	Default string // default value (default tag)
}

// collectConfigFields walks t recursively and collects all configurable leaf fields.
// envPrefix is accumulated from parent struct envPrefix tags (caarlos0/env convention).
func collectConfigFields(t reflect.Type, envPrefix string) []fieldDesc {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var result []fieldDesc
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		ft := sf.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct {
			// Recurse into nested struct, accumulating envPrefix from the parent tag.
			childPrefix := envPrefix + sf.Tag.Get("envPrefix")
			result = append(result, collectConfigFields(ft, childPrefix)...)
			continue
		}

		envTag := sf.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Derive display name: prefer field tag, then json tag, then Go field name.
		name := sf.Tag.Get("field")
		if name == "" {
			name = strings.Split(sf.Tag.Get("json"), ",")[0]
		}
		if name == "" || name == "-" {
			name = strings.ToLower(sf.Name)
		}

		result = append(result, fieldDesc{
			Name:    name,
			Flag:    sf.Tag.Get("cli"),
			EnvVar:  envPrefix + envTag,
			Default: sf.Tag.Get("default"),
		})
	}
	return result
}

// printCommandUsage prints command name, description, and an aligned options table
// derived by reflecting over the zero value of T.
func printCommandUsage[T any](w io.Writer, name, helpDesc string) {
	_, _ = fmt.Fprintf(w, "Command: %s\n", name)
	if helpDesc != "" {
		_, _ = fmt.Fprintf(w, "  %s\n", helpDesc)
	}
	_, _ = fmt.Fprintln(w)

	var zero T
	fields := collectConfigFields(reflect.TypeOf(zero), "")
	if len(fields) == 0 {
		return
	}

	// Compute column widths.
	nameW, flagW, envW := len("Name"), len("Flag"), len("Env")
	for _, f := range fields {
		if n := len(f.Name); n > nameW {
			nameW = n
		}
		if f.Flag != "" {
			if n := len(f.Flag) + 2; n > flagW { // +2 for "--"
				flagW = n
			}
		}
		if n := len(f.EnvVar); n > envW {
			envW = n
		}
	}

	sep := strings.Repeat("─", nameW+flagW+envW+14)
	_, _ = fmt.Fprintf(w, "Options:\n")
	_, _ = fmt.Fprintf(w, "  %-*s  %-*s  %-*s  %s\n", nameW, "Name", flagW, "Flag", envW, "Env", "Default")
	_, _ = fmt.Fprintf(w, "  %s\n", sep)

	for _, f := range fields {
		flag := ""
		if f.Flag != "" {
			flag = "--" + f.Flag
		}
		def := ""
		if f.Default != "" {
			def = `"` + f.Default + `"`
		}
		_, _ = fmt.Fprintf(w, "  %-*s  %-*s  %-*s  %s\n",
			nameW, f.Name,
			flagW, flag,
			envW, f.EnvVar,
			def,
		)
	}
	_, _ = fmt.Fprintln(w)
}
