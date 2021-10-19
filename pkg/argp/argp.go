package argp

import (
	"fmt"
	"strconv"
	"strings"
)

type SubcommandGroup struct {
	subcommands map[string]*Subcommand
}

func CreateSubcommandGroup() *SubcommandGroup {
	return &SubcommandGroup{
		subcommands: make(map[string]*Subcommand),
	}
}

func (sg *SubcommandGroup) ParseSubcommand(args []string) (*Subcommand, error) {
	if len(args) < 2 {
		fmt.Println("A subcommand is required.")
		sg.PrintUsage("")
		return nil, fmt.Errorf("no subcommand speicified")
	}

	if args[1] == "--help" || args[1] == "-h" {
		sg.PrintUsage("")
		return nil, nil
	}

	subcommand, exists := sg.subcommands[strings.ToLower(args[1])]
	if !exists {
		fmt.Printf("Unknown subcommand %s.\n", args[1])
		sg.PrintUsage("")
		return nil, fmt.Errorf("unknown subcommand %s", args[1])
	}

	err := subcommand.ParseFlags(args[2:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments of subcommand %s: %w", subcommand.Name, err)
	}
	return subcommand, nil
}

func (sg *SubcommandGroup) PrintUsage(indent string) {
	fmt.Printf("%sAvailable subcommands:\n", indent)
	indent += "  "
	for _, subcommand := range sg.subcommands {
		fmt.Printf("%s%s: %s\n", indent, subcommand.Name, subcommand.Usage)
		subcommand.PrintUsage(indent + "  ")
	}
}

func (sg *SubcommandGroup) AddSubcommand(subcommand *Subcommand) *Subcommand {
	sg.subcommands[strings.ToLower(subcommand.Name)] = subcommand
	return subcommand
}

type Subcommand struct {
	Name  string
	Usage string
	FlagSet
}

func CreateSubcommand(name string, usage string) *Subcommand {
	return &Subcommand{
		Name: name, Usage: usage, FlagSet: *CreateFlagSet(),
	}
}

type FlagSet struct {
	flags      map[string]*Flag
	positional []*Flag
}

func CreateFlagSet() *FlagSet {
	return &FlagSet{
		flags: make(map[string]*Flag, 10),
	}
}

func (flagSet *FlagSet) ParseFlags(args []string) error {
	currentFlag := (*Flag)(nil)
	pos := 0
	for _, arg := range args {
		if len(arg) > 1 && arg[0] == '-' {
			stripped := arg[1:]
			if stripped[0] == '-' {
				stripped = stripped[1:]
			}

			flag, exists := flagSet.flags[stripped]
			if !exists {
				if stripped == "h" || stripped == "help" {
					flagSet.PrintUsage("")
					return nil
				}
				return flagSet.handleError("unknown flag %s", arg)
			}
			if currentFlag != nil {
				boolFlagValue, success := currentFlag.Value.(*boolValue)
				if success {
					*boolFlagValue = true
					currentFlag = flag
					continue
				}
				return flagSet.handleError("unknown flag %s in value of other flag %s", arg, currentFlag.Names[0])
			}
			currentFlag = flag
			continue
		}

		if currentFlag != nil {
			err := currentFlag.Value.FromString(arg)
			if err != nil {
				return flagSet.handleError("failed to parse value for flag %s = %s: %v", currentFlag.Names[0], arg, err)
			}
			currentFlag = nil
			continue
		}

		if len(flagSet.positional) <= pos {
			return flagSet.handleError("Encountered additional positional argument %s", arg)
		}

		err := flagSet.positional[pos].Value.FromString(arg)
		pos++
		if err != nil {
			return flagSet.handleError("failed to parse positional argument %s = %s: %v", currentFlag.Names[0], arg, err)
		}
	}
	return nil
}

func (flagSet *FlagSet) handleError(format string, args ...interface{}) error {
	cappedFormat := strings.ToUpper(string(format[0])) + format[1:] + "\n"
	fmt.Printf(cappedFormat, args)
	flagSet.PrintUsage("")
	return fmt.Errorf(format+"\n", args)
}

func (flagSet *FlagSet) PrintUsage(indent string) {
	if len(flagSet.positional) > 0 {
		fmt.Printf("%sPositional arguments:\n", indent)
		for _, flag := range flagSet.positional {
			fmt.Printf("%s  %10s: %s\n", indent, flag.Names[0], flag.Usage)
			if len(*flag.DefaultText) > 0 {
				fmt.Printf("%s            Default: %s\n", indent, *flag.DefaultText)
			}
		}
	}
	if len(flagSet.flags) > 0 {
		fmt.Printf("%sFlags:\n", indent)

		for _, flag := range flagSet.flags {
			flagVariants := make([]string, len(flag.Names))
			for i, name := range flag.Names {
				if len(name) == 1 {
					flagVariants[i] = "-" + name
				} else {
					flagVariants[i] = "--" + name
				}
			}
			fmt.Printf("%s  %s:\n", indent, strings.Join(flagVariants, ", "))
			fmt.Printf("%s    %s\n", indent, flag.Usage)
			fmt.Printf("%s    Default: %s\n", indent, *flag.DefaultText)
		}
	}
}

type FlagBuildArgs struct {
	Names       []string
	Usage       string
	DefaultText *string
}

type Flag struct {
	FlagBuildArgs
	Default FlagValue
	Value   FlagValue
}

type FlagValue interface {
	String() string
	FromString(text string) error
}

func (flagSet *FlagSet) Bool(flagArgs FlagBuildArgs, defaultValue bool) *bool {
	value := boolValue(defaultValue)
	_defaultValue := boolValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*bool)(&value)
}

func (flagSet *FlagSet) Int(flagArgs FlagBuildArgs, defaultValue int) *int {
	value := intValue(defaultValue)
	_defaultValue := intValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*int)(&value)
}

func (flagSet *FlagSet) Uint(flagArgs FlagBuildArgs, defaultValue uint) *uint {
	value := uintValue(defaultValue)
	_defaultValue := uintValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*uint)(&value)
}

func (flagSet *FlagSet) String(flagArgs FlagBuildArgs, defaultValue string) *string {
	value := stringValue(defaultValue)
	_defaultValue := stringValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*string)(&value)
}

func (flagSet *FlagSet) PositionalBool(flagArgs FlagBuildArgs, defaultValue bool) *bool {
	value := boolValue(defaultValue)
	_defaultValue := boolValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*bool)(&value)
}

func (flagSet *FlagSet) PositionalInt(flagArgs FlagBuildArgs, defaultValue int) *int {
	value := intValue(defaultValue)
	_defaultValue := intValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*int)(&value)
}

func (flagSet *FlagSet) PositionalUint(flagArgs FlagBuildArgs, defaultValue uint) *uint {
	value := uintValue(defaultValue)
	_defaultValue := uintValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*uint)(&value)
}

func (flagSet *FlagSet) PositionalString(flagArgs FlagBuildArgs, defaultValue string) *string {
	value := stringValue(defaultValue)
	_defaultValue := stringValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*string)(&value)
}

func (flagSet *FlagSet) addFlag(flag *Flag) {
	if flag.DefaultText == nil {
		defText := flag.Default.String()
		flag.DefaultText = &defText
	}
	for _, name := range flag.Names {
		flagSet.flags[name] = flag
	}
}

func (flagSet *FlagSet) addPositional(flag *Flag) {
	if flag.DefaultText == nil {
		defText := flag.Default.String()
		flag.DefaultText = &defText
	}
	flagSet.positional = append(flagSet.positional, flag)
}

type boolValue bool

func (value *boolValue) String() string {
	return strconv.FormatBool(bool(*value))
}

func (value *boolValue) FromString(text string) error {
	parsed, err := strconv.ParseBool(text)
	if err != nil {
		return err
	}
	*value = boolValue(parsed)
	return nil
}

type uintValue uint

func (value *uintValue) String() string {
	return strconv.FormatUint(uint64(*value), 10)
}

func (value *uintValue) FromString(text string) error {
	parsed, err := strconv.ParseUint(text, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*value = uintValue(parsed)
	return nil
}

type intValue int

func (value *intValue) String() string {
	return strconv.FormatInt(int64(*value), 10)
}

func (value *intValue) FromString(text string) error {
	parsed, err := strconv.ParseInt(text, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*value = intValue(parsed)
	return nil
}

type stringValue string

func (value *stringValue) String() string {
	return string(*value)
}

func (value *stringValue) FromString(text string) error {
	*value = stringValue(text)
	return nil
}
