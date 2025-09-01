package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// FlagDef describes a configuration item that can be populated from
// a command-line flag, environment variable and YAML configuration file.
type FlagDef struct {
	// Name is the command line flag name, e.g. "log-level".
	Name string
	// EnvName is the suffix of the environment variable (without the
	// OPTIMUM_ prefix). If empty, it is derived from Name.
	EnvName string
	// Usage describes the flag for help output.
	Usage string
	// Value points to the field in the configuration struct that should be populated.
	Value interface{}
}

// Load populates cfg using the provided flag definitions. Configuration values
// are loaded from, in ascending order of priority: YAML file, environment
// variables and command-line flags. The YAML file path is resolved from the
// --config flag which defaults to "config.yaml".
func Load(cfg interface{}, flagDefs []FlagDef) error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var confPath string
	fs.StringVar(&confPath, "config", "config.yaml", "path to YAML config file")

	// Temporary storage for parsed flag values and the set of flags specified
	// by the user.
	parsed := make([]interface{}, len(flagDefs))
	defaults := make([]interface{}, len(flagDefs))

	for i, fd := range flagDefs {
		switch fd.Value.(type) {
		case *string:
			var v string
			fs.StringVar(&v, fd.Name, "", fd.Usage)
			parsed[i] = &v
			defaults[i] = ""
		case *int:
			var v int
			fs.IntVar(&v, fd.Name, 0, fd.Usage)
			parsed[i] = &v
			defaults[i] = 0
		case *bool:
			var v bool
			fs.BoolVar(&v, fd.Name, false, fd.Usage)
			parsed[i] = &v
			defaults[i] = false
		case *float64:
			var v float64
			fs.Float64Var(&v, fd.Name, 0, fd.Usage)
			parsed[i] = &v
			defaults[i] = float64(0)
		case *float32:
			var v float64
			fs.Float64Var(&v, fd.Name, 0, fd.Usage)
			parsed[i] = &v
			defaults[i] = float64(0)
		case *[]string:
			var v []string
			fs.Var(&stringSliceValue{val: &v}, fd.Name, fd.Usage)
			parsed[i] = &v
			defaults[i] = []string(nil)
		default:
			return fmt.Errorf("unsupported flag type %T", fd.Value)
		}
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	changed := make([]bool, len(flagDefs))
	for i, fd := range flagDefs {
		switch fd.Value.(type) {
		case *string:
			changed[i] = *parsed[i].(*string) != defaults[i].(string)
		case *int:
			changed[i] = *parsed[i].(*int) != defaults[i].(int)
		case *bool:
			changed[i] = *parsed[i].(*bool) != defaults[i].(bool)
		case *float64:
			changed[i] = *parsed[i].(*float64) != defaults[i].(float64)
		case *float32:
			changed[i] = float64(*parsed[i].(*float64)) != defaults[i].(float64)
		case *[]string:
			changed[i] = len(*parsed[i].(*[]string)) != 0
		}
	}

	// Load YAML file if present.
	if data, err := os.ReadFile(confPath); err == nil { //nolint:gosec // config path is user supplied
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Overlay environment variables.
	for _, fd := range flagDefs {
		env := fd.EnvName
		if env == "" {
			env = strings.ToUpper(strings.ReplaceAll(fd.Name, "-", "_"))
		}
		env = "OPTIMUM_" + env
		if val, ok := os.LookupEnv(env); ok {
			if err := setValue(fd.Value, val); err != nil {
				return fmt.Errorf("invalid value for %s: %w", env, err)
			}
		}
	}

	// Finally overlay flags that were explicitly set.
	for i, fd := range flagDefs {
		if !changed[i] {
			continue
		}
		switch ptr := fd.Value.(type) {
		case *string:
			*ptr = *parsed[i].(*string)
		case *int:
			*ptr = *parsed[i].(*int)
		case *bool:
			*ptr = *parsed[i].(*bool)
		case *float64:
			*ptr = *parsed[i].(*float64)
		case *float32:
			*ptr = float32(*parsed[i].(*float64))
		case *[]string:
			*ptr = *parsed[i].(*[]string)
		default:
			return fmt.Errorf("unsupported flag type %T", fd.Value)
		}
	}

	return nil
}

func setValue(dest interface{}, val string) error {
	switch ptr := dest.(type) {
	case *string:
		*ptr = val
	case *int:
		i, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*ptr = i
	case *bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		*ptr = b
	case *float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		*ptr = f
	case *float32:
		f, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		*ptr = float32(f)
	case *[]string:
		if val == "" {
			*ptr = []string{}
		} else {
			*ptr = strings.Split(val, ",")
		}
	default:
		return fmt.Errorf("unsupported type %T", dest)
	}
	return nil
}

// stringSliceValue implements flag.Value for []string to enable command line
// parsing of comma separated lists.
type stringSliceValue struct {
	val *[]string
}

func (s *stringSliceValue) String() string {
	if s.val == nil {
		return ""
	}
	return strings.Join(*s.val, ",")
}

func (s *stringSliceValue) Set(value string) error {
	if value == "" {
		*s.val = []string{}
	} else {
		*s.val = strings.Split(value, ",")
	}
	return nil
}
