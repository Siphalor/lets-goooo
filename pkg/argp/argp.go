package argp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SubcommandGroup groups multiple subcommands together into a group that can be evaluated on its own.
type SubcommandGroup struct {
	// subcommands is a map from the identifiers (names) of the subcommands to said subcommands
	subcommands map[string]*Subcommand
}

// CreateSubcommandGroup creates a new subcommand group.
func CreateSubcommandGroup() *SubcommandGroup {
	return &SubcommandGroup{
		subcommands: make(map[string]*Subcommand),
	}
}

// ParseSubcommand parses the given arguments and resolves the called subcommand.
func (sg *SubcommandGroup) ParseSubcommand(args []string) (*Subcommand, error) {
	if len(args) < 2 { // The first arg is always the call to the executable, so minimum of two
		fmt.Println("A subcommand is required.")
		sg.PrintUsage("")
		return nil, fmt.Errorf("no subcommand speicified")
	}

	if args[1] == "--help" || args[1] == "-h" {
		sg.PrintUsage("")
		return nil, nil
	}

	// Look for a matching subcommand (case-insensitive)
	subcommand, exists := sg.subcommands[strings.ToLower(args[1])]
	if !exists {
		fmt.Printf("Unknown subcommand %s.\n", args[1])
		sg.PrintUsage("")
		return nil, fmt.Errorf("unknown subcommand %s", args[1])
	}

	err := subcommand.ParseFlags(args[2:]) // Parse all further args for the subcommand
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments of subcommand %s: %w", subcommand.Name, err)
	}
	return subcommand, nil
}

// PrintUsage prints the usage information for all subcommands.
func (sg *SubcommandGroup) PrintUsage(indent string) {
	fmt.Printf("%sAvailable subcommands:\n", indent)
	indent += "  "
	for _, subcommand := range sg.subcommands {
		fmt.Printf("%s%s: %s\n", indent, subcommand.Name, subcommand.Usage)
		subcommand.PrintUsage(indent + "  ")
	}
}

// AddSubcommand adds a subcommand to the command group.
// It returns the given subcommand to allow easy assignment.
func (sg *SubcommandGroup) AddSubcommand(subcommand *Subcommand) *Subcommand {
	sg.subcommands[strings.ToLower(subcommand.Name)] = subcommand
	return subcommand
}

// Subcommand is the representation of a subcommand, with name usage information and flags.
type Subcommand struct {
	// Name is the name with which the subcommand is called
	Name string
	// Usage provides a description of what the subcommand does
	Usage string
	FlagSet
}

// CreateSubcommand creates a new subcommand with the given name and usage information.
func CreateSubcommand(name string, usage string) *Subcommand {
	return &Subcommand{
		Name: name, Usage: usage, FlagSet: *CreateFlagSet(),
	}
}

// FlagSet groups together flags and positional arguments.
type FlagSet struct {
	// flags is a map from the unique flag names to the actual Flag structs
	flags map[string]*Flag
	// positional defines the positional arguments in that exact order
	positional []*Flag
}

// CreateFlagSet creates a new, empty flag set.
func CreateFlagSet() *FlagSet {
	return &FlagSet{
		flags: make(map[string]*Flag, 10),
	}
}

// ParseFlags parses the flags for current FlagSet.
func (flagSet *FlagSet) ParseFlags(args []string) error {
	currentFlag := (*Flag)(nil) // The current flag - used to resolve values that follow a flag
	pos := 0                    // The current index on positional arguments
	for _, arg := range args {
		if len(arg) > 1 && arg[0] == '-' { // Flag argument
			stripped := arg[1:]
			if stripped[0] == '-' { // Handle double dashes
				stripped = stripped[1:]
			}
			stripped = strings.ToLower(stripped) // Flags are case-insensitive

			flag, exists := flagSet.flags[stripped]
			if !exists { // No such flag exists
				if stripped == "h" || stripped == "help" { // Special case the help flags
					flagSet.PrintUsage("")
					os.Exit(0)
				}
				return flagSet.handleError("unknown flag %s", arg)
			}
			if currentFlag != nil { // Found a flag while looking for a value of a previous flag
				return flagSet.handleError("unknown flag %s in value of other flag %s", arg, currentFlag.Names[0])
			}

			if flag.TakesValue() { // The parsed flag requires a following value
				currentFlag = flag
			} else { // The flag doesn't take an argument - this is currently special cased for bools
				boolFlagValue, isBool := flag.Value.(*boolValue)
				if isBool {
					*boolFlagValue = true
				}
				currentFlag = nil // Reset the value awaiter - just in case
			}
			continue // Flag handling done
		}

		if currentFlag != nil { // Check if the last flag requires a trailing value
			err := currentFlag.Value.FromString(arg)
			if err != nil {
				return flagSet.handleError("failed to parse value for flag %s = %s: %v", currentFlag.Names[0], arg, err)
			}
			currentFlag = nil // No more values
			continue
		}

		// Check if there is still room for positional arguments
		if len(flagSet.positional) <= pos {
			return flagSet.handleError("Encountered additional positional argument %s", arg)
		}

		// Try to parse as the next positional argument
		err := flagSet.positional[pos].Value.FromString(arg)
		pos++
		if err != nil {
			return flagSet.handleError("failed to parse positional argument %s = %s: %v", currentFlag.Names[0], arg, err)
		}
	}
	return nil
}

// handleError handles a parse error.
func (flagSet *FlagSet) handleError(format string, args ...interface{}) error {
	// The user facing value should be capitalized, where Errorf should begin with a lowercase letter
	cappedFormat := strings.ToUpper(string(format[0])) + format[1:] + "\n"
	fmt.Printf(cappedFormat, args...)
	flagSet.PrintUsage("")
	return fmt.Errorf(format+"\n", args...)
}

// PrintUsage prints usage information for the FlagSet.
func (flagSet *FlagSet) PrintUsage(indent string) {
	// Positional arguments, if any, go first
	if len(flagSet.positional) > 0 {
		fmt.Printf("%sPositional arguments:\n", indent)
		for _, flag := range flagSet.positional {
			fmt.Printf("%s  %10s: %s\n", indent, flag.Names[0], flag.Usage)
			// Only show default text if it's not empty - looks ugly otherwise
			if *flag.DefaultText != "" {
				fmt.Printf("%s            Default: %s\n", indent, *flag.DefaultText)
			}
		}
	}
	// Flags, if any, come next
	if len(flagSet.flags) > 0 {
		fmt.Printf("%sFlags:\n", indent)

		// The unique flags map is required because some flags are registered under multiple names
		uniqueFlags := make(map[*Flag]bool, len(flagSet.flags))
		for _, flag := range flagSet.flags {
			_, exists := uniqueFlags[flag] // check if the flag has been printed already
			if exists {
				continue // if so, continue
			}
			uniqueFlags[flag] = true

			// Collect all flag names (aliases)
			flagVariants := make([]string, len(flag.Names))
			for i, name := range flag.Names {
				if len(name) == 1 { // Single letter names are usually used with a single dash
					flagVariants[i] = "-" + name
				} else { // Longer names usually (Linux) use double dashes (both are allowed though)
					flagVariants[i] = "--" + name
				}
			}

			takesValue := flag.TakesValue() // Check if the flag requires a trailing value
			if takesValue {                 // If so, write an according usage information
				fmt.Printf("%s  %s <value>:\n", indent, strings.Join(flagVariants, ", "))
			} else {
				fmt.Printf("%s  %s:\n", indent, strings.Join(flagVariants, ", "))
			}
			fmt.Printf("%s    %s\n", indent, flag.Usage)
			// Only show default text if it's not empty and if it requires a value - looks ugly otherwise
			if takesValue && *flag.DefaultText != "" {
				fmt.Printf("%s    Default: %s\n", indent, *flag.DefaultText)
			}
		}
	}
}

// FlagBuildArgs collects construction arguments for a Flag.
type FlagBuildArgs struct {
	// Names defines all the names (aliases) for a flag.
	// Positional arguments only use this in the usage information.
	Names []string
	// Usage is a description of what the Flag does
	Usage string
	// DefaultText is an override for Default.String() to allow special casing
	DefaultText *string
}

// Flag represents the metadata and values of a flag.
type Flag struct {
	FlagBuildArgs
	// Default is the default value of the flag
	Default FlagValue
	// Value is the actual value of the flag
	Value FlagValue
}

// TakesValue returns, whether the flag requires a trailing value argument
func (flag *Flag) TakesValue() bool {
	boolDefault, isBool := flag.Default.(*boolValue) // Check if it's a bool
	// If it's a bool with false as default, then no value is required
	return !isBool || bool(*boolDefault)
}

// FlagValue can be used in a Flag to parse/serialize values.
type FlagValue interface {
	// String brings the value into string representation
	String() string
	// FromString parses the user input
	FromString(text string) error
}

// ARGUMENT CONSTRUCTOR METHODS
// this could be greatly simplified if Go had native generics/templates

// Bool creates a bool argument.
func (flagSet *FlagSet) Bool(flagArgs FlagBuildArgs, defaultValue bool) *bool {
	value := boolValue(defaultValue)
	_defaultValue := boolValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*bool)(&value)
}

// Int creates an int argument.
func (flagSet *FlagSet) Int(flagArgs FlagBuildArgs, defaultValue int) *int {
	value := intValue(defaultValue)
	_defaultValue := intValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*int)(&value)
}

// Uint creates an uint argument.
func (flagSet *FlagSet) Uint(flagArgs FlagBuildArgs, defaultValue uint) *uint {
	value := uintValue(defaultValue)
	_defaultValue := uintValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*uint)(&value)
}

// String creates a string argument.
func (flagSet *FlagSet) String(flagArgs FlagBuildArgs, defaultValue string) *string {
	value := stringValue(defaultValue)
	_defaultValue := stringValue(defaultValue)
	flagSet.addFlag(&Flag{flagArgs, &_defaultValue, &value})
	return (*string)(&value)
}

// PositionalBool creates a positional bool argument.
func (flagSet *FlagSet) PositionalBool(flagArgs FlagBuildArgs, defaultValue bool) *bool {
	value := boolValue(defaultValue)
	_defaultValue := boolValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*bool)(&value)
}

// PositionalInt creates a positional int argument.
func (flagSet *FlagSet) PositionalInt(flagArgs FlagBuildArgs, defaultValue int) *int {
	value := intValue(defaultValue)
	_defaultValue := intValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*int)(&value)
}

// PositionalUint creates a positional uint argument.
func (flagSet *FlagSet) PositionalUint(flagArgs FlagBuildArgs, defaultValue uint) *uint {
	value := uintValue(defaultValue)
	_defaultValue := uintValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*uint)(&value)
}

// PositionalString creates a positional string argument.
func (flagSet *FlagSet) PositionalString(flagArgs FlagBuildArgs, defaultValue string) *string {
	value := stringValue(defaultValue)
	_defaultValue := stringValue(defaultValue)
	flagSet.addPositional(&Flag{flagArgs, &_defaultValue, &value})
	return (*string)(&value)
}

// addFlag adds a new flag to the FlagSet.
func (flagSet *FlagSet) addFlag(flag *Flag) {
	if flag.DefaultText == nil {
		defText := flag.Default.String()
		flag.DefaultText = &defText
	}
	for _, name := range flag.Names {
		flagSet.flags[name] = flag
	}
}

// addPositional adds a new positional argument.
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
