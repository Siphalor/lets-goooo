// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package argp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//  ____        _                                                   _
// / ___| _   _| |__   ___ ___  _ __ ___  _ __ ___   __ _ _ __   __| |___
// \___ \| | | | '_ \ / __/ _ \| '_ ` _ \| '_ ` _ \ / _` | '_ \ / _` / __|
//  ___) | |_| | |_) | (_| (_) | | | | | | | | | | | (_| | | | | (_| \__ \
// |____/ \__,_|_.__/ \___\___/|_| |_| |_|_| |_| |_|\__,_|_| |_|\__,_|___/

func TestSubcommandGroup_ParseSubcommand(t *testing.T) {
	sg := CreateSubcommandGroup()
	sub1 := CreateSubcommand("sub1", "sub1 usage")
	sub2 := CreateSubcommand("sub2", "sub2 usage")
	sg.AddSubcommand(sub1)
	sg.AddSubcommand(sub2)

	if sub, err := sg.ParseSubcommand([]string{"sub1"}); assert.NoError(t, err) {
		assert.EqualValues(t, sub1, sub)
	}
	if sub, err := sg.ParseSubcommand([]string{"sub2"}); assert.NoError(t, err) {
		assert.EqualValues(t, sub2, sub)
	}

	//  ___
	// | __|_ _ _ _ ___ _ _ ___
	// | _|| '_| '_/ _ \ '_(_-<
	// |___|_| |_| \___/_| /__/

	_, err := sg.ParseSubcommand([]string{})
	assert.EqualError(t, err, "no subcommand specified")
	_, err = sg.ParseSubcommand([]string{"unknown"})
	assert.EqualError(t, err, "unknown subcommand \"unknown\"")
	_, err = sg.ParseSubcommand([]string{"sub1", "--wtf"})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to parse arguments of subcommand sub1")
	}
}

func ExampleSubcommandGroup_PrintUsage() {
	sg := CreateSubcommandGroup()
	sub1 := CreateSubcommand("sub1", "sub1 usage")
	sub2 := CreateSubcommand("sub2", "sub2 usage\nmultiline")
	sg.AddSubcommand(sub1)
	sg.AddSubcommand(sub2)

	sg.PrintUsage("  ")

	// Output:
	//   Available subcommands:
	//     sub1:
	//       sub1 usage
	//     sub2:
	//       sub2 usage
	//       multiline
}

func TestSubcommandGroup_AddSubcommand(t *testing.T) {
	sg := CreateSubcommandGroup()
	sg.AddSubcommand(&Subcommand{
		Name: "test",
	})
	sg.AddSubcommand(&Subcommand{
		Name: "WeIrD-CaSe",
	})
	assert.Len(t, sg.subcommands, 2)
	assert.Contains(t, sg.subcommands, "test")
	assert.Contains(t, sg.subcommands, "weird-case")
}
