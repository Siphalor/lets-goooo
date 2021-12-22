package argp

import (
	"fmt"
	"strings"
)

// SubcommandGroup groups multiple subcommands together into a group that can be evaluated on its own.
type SubcommandGroup struct {
	// subcommands is a map from the identifiers (names) of the subcommands to said subcommands
	subcommands map[string]*Subcommand
	// orderedSubcommands is a companion to subcommands to retain the order
	orderedSubcommands []*Subcommand
}

// CreateSubcommandGroup creates a new subcommand group.
func CreateSubcommandGroup() *SubcommandGroup {
	return &SubcommandGroup{
		subcommands: make(map[string]*Subcommand),
	}
}

// ParseSubcommand parses the given arguments and resolves the called subcommand.
func (sg *SubcommandGroup) ParseSubcommand(args []string) (*Subcommand, error) {
	if len(args) < 1 { // The first arg is always the call to the executable, so minimum of two
		fmt.Println("A subcommand is required.")
		sg.PrintUsage("")
		return nil, fmt.Errorf("no subcommand specified")
	}

	if args[0] == "--help" || args[0] == "-h" {
		sg.PrintUsage("")
		return nil, nil
	}

	// Look for a matching subcommand (case-insensitive)
	subcommand, exists := sg.subcommands[strings.ToLower(args[0])]
	if !exists {
		fmt.Printf("Unknown subcommand %s.\n", args[0])
		sg.PrintUsage("")
		return nil, fmt.Errorf("unknown subcommand \"%s\"", args[0])
	}

	err := subcommand.ParseFlags(args[1:]) // Parse all further args for the subcommand
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments of subcommand %s: %w", subcommand.Name, err)
	}
	return subcommand, nil
}

// PrintUsage prints the usage information for all subcommands.
func (sg *SubcommandGroup) PrintUsage(indent string) {
	fmt.Printf("%sAvailable subcommands:\n", indent)
	indent += "  "
	for _, subcommand := range sg.orderedSubcommands {
		fmt.Printf("%s%s:\n", indent, subcommand.Name)
		for _, line := range strings.Split(subcommand.Usage, "\n") {
			fmt.Printf("%s  %s\n", indent, line)
		}
		subcommand.PrintUsage(indent + "  ")
	}
}

// AddSubcommand adds a subcommand to the command group.
// It returns the given subcommand to allow easy assignment.
func (sg *SubcommandGroup) AddSubcommand(subcommand *Subcommand) *Subcommand {
	sg.subcommands[strings.ToLower(subcommand.Name)] = subcommand
	sg.orderedSubcommands = append(sg.orderedSubcommands, subcommand)
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
