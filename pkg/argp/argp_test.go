package argp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//  _____ _             ____       _
// |  ___| | __ _  __ _/ ___|  ___| |_   _ __   __ _ _ __ ___  ___
// | |_  | |/ _` |/ _` \___ \ / _ \ __| | '_ \ / _` | '__/ __|/ _ \
// |  _| | | (_| | (_| |___) |  __/ |_  | |_) | (_| | |  \__ \  __/
// |_|   |_|\__,_|\__, |____/ \___|\__| | .__/ \__,_|_|  |___/\___|
//                |___/                 |_|

func TestFlagSet_ParseFlags(t *testing.T) {
	fs := CreateFlagSet()

	resetFlags := func() {
		for _, flag := range fs.orderedFlags {
			_ = flag.Value.FromString(flag.Default.String())
		}
		for _, flag := range fs.positional {
			_ = flag.Value.FromString(flag.Default.String())
		}
	}

	pos1 := fs.PositionalInt(FlagBuildArgs{
		Names: []string{"pos1"},
	}, 123)
	pos2 := fs.PositionalString(FlagBuildArgs{
		Names: []string{"pos2"},
	}, "")
	flagAlpha := fs.Int(FlagBuildArgs{
		Names: []string{"alpha", "a"},
	}, 123)
	flagNoVal := fs.Bool(FlagBuildArgs{
		Names: []string{"no-val"},
	}, false)

	// no flags passed
	if assert.NoError(t, fs.ParseFlags([]string{})) {
		assert.Equal(t, 123, *pos1)
		assert.Equal(t, "", *pos2)
		assert.Equal(t, 123, *flagAlpha)
		assert.Equal(t, false, *flagNoVal)
	}
	resetFlags()

	// mixed
	if assert.NoError(t, fs.ParseFlags([]string{"100", "abc", "-a", "100", "--no-val"})) {
		assert.Equal(t, 100, *pos1)
		assert.Equal(t, "abc", *pos2)
		assert.Equal(t, 100, *flagAlpha)
		assert.Equal(t, true, *flagNoVal)
	}
	resetFlags()

	// long flag
	if assert.NoError(t, fs.ParseFlags([]string{"--alpha", "456"})) {
		assert.Equal(t, 123, *pos1)
		assert.Equal(t, "", *pos2)
		assert.Equal(t, 456, *flagAlpha)
		assert.Equal(t, false, *flagNoVal)
	}
	resetFlags()

	// split positional args
	if assert.NoError(t, fs.ParseFlags([]string{"456", "--alpha", "456", "abc"})) {
		assert.Equal(t, 456, *pos1)
		assert.Equal(t, "abc", *pos2)
		assert.Equal(t, 456, *flagAlpha)
		assert.Equal(t, false, *flagNoVal)
	}
	resetFlags()

	// unusual dashes
	if assert.NoError(t, fs.ParseFlags([]string{"-no-val", "--a", "101"})) {
		assert.Equal(t, 123, *pos1)
		assert.Equal(t, "", *pos2)
		assert.Equal(t, 101, *flagAlpha)
		assert.Equal(t, true, *flagNoVal)
	}
	resetFlags()

	//  ___
	// | __|_ _ _ _ ___ _ _ ___
	// | _|| '_| '_/ _ \ '_(_-<
	// |___|_| |_| \___/_| /__/

	// unknown flag
	assert.EqualError(t, fs.ParseFlags([]string{"--unknown"}), "unknown flag \"--unknown\"")
	resetFlags()
	// flag in flag value
	assert.EqualError(t, fs.ParseFlags([]string{"--alpha", "--no-val"}), "unknown flag \"--no-val\" in value of other flag \"--alpha\"")
	// too many positional args
	assert.EqualError(t, fs.ParseFlags([]string{"100", "text", "wtf"}), "encountered additional positional argument \"wtf\"")
	resetFlags()
	// invalid positional arg
	if err := fs.ParseFlags([]string{"error"}); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to parse positional argument \"pos2\" = \"error\"")
	}
	resetFlags()
	// invalid flag arg
	if err := fs.ParseFlags([]string{"--alpha", "error"}); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to parse value for flag \"alpha\" = \"error\"")
	}
	// positional with no value
	assert.EqualError(t, fs.ParseFlags([]string{"--alpha"}), "trailing value is missing for argument \"alpha\"")
}

//  _____ _             ____       _
// |  ___| | __ _  __ _/ ___|  ___| |_   _   _ ___  __ _  __ _  ___
// | |_  | |/ _` |/ _` \___ \ / _ \ __| | | | / __|/ _` |/ _` |/ _ \
// |  _| | | (_| | (_| |___) |  __/ |_  | |_| \__ \ (_| | (_| |  __/
// |_|   |_|\__,_|\__, |____/ \___|\__|  \__,_|___/\__,_|\__, |\___|
//                |___/                                  |___/

func ExampleFlagSet_PrintUsage_positionalArgs() {
	fs := CreateFlagSet()
	fs.PositionalInt(FlagBuildArgs{
		Names: []string{"arg1"},
		Usage: "Hello Arg1!",
	}, 123)
	arg2def := "0x101"
	fs.PositionalUint(FlagBuildArgs{
		Names:       []string{"arg2has-a-long-name", "arg2alt"},
		DefaultText: &arg2def,
	}, 257)
	fs.PositionalString(FlagBuildArgs{
		Names: []string{},
		Usage: "A multiline\ndescription",
	}, "")

	fs.PrintUsage("  ")

	// Output:
	//   Positional arguments:
	//     arg1: Hello Arg1!
	//         Default: 123
	//     arg2has-a-long-name:
	//         Default: 0x101
	//     <arg>: A multiline
	//         description
}

func ExampleFlagSet_PrintUsage_flagArgs() {
	fs := CreateFlagSet()
	fs.Int(FlagBuildArgs{
		Names: []string{"arg1", "1"},
		Usage: "This is how to use arg1 :)",
	}, 123)
	fs.Bool(FlagBuildArgs{
		Names: []string{"arg2", "2"},
		Usage: "A classic flag argument",
	}, false)
	fs.String(FlagBuildArgs{
		Names: []string{"3"},
		Usage: "a multiline\ndescription",
	}, "")

	fs.PrintUsage("  ")

	// Output:
	//   Flags:
	//     --arg1, -1 <value>:
	//         This is how to use arg1 :)
	//         Default: 123
	//     --arg2, -2:
	//         A classic flag argument
	//     -3 <value>:
	//         a multiline
	//         description
}

func ExampleFlagSet_PrintUsage_mixedArgs() {
	fs := CreateFlagSet()
	fs.PositionalInt(FlagBuildArgs{
		Names: []string{"parg1"},
		Usage: "This is positional arg 1",
	}, 100)
	fs.Int(FlagBuildArgs{
		Names: []string{"farg1"},
		Usage: "This is flag arg 1",
	}, 123)

	fs.PrintUsage("")

	// Output:
	// Positional arguments:
	//   parg1: This is positional arg 1
	//       Default: 100
	// Flags:
	//   --farg1 <value>:
	//       This is flag arg 1
	//       Default: 123
}

//  _____ _               _   _
// |  ___| | __ _  __ _  | \ | | __ _ _ __ ___   ___
// | |_  | |/ _` |/ _` | |  \| |/ _` | '_ ` _ \ / _ \
// |  _| | | (_| | (_| | | |\  | (_| | | | | | |  __/
// |_|   |_|\__,_|\__, | |_| \_|\__,_|_| |_| |_|\___|
//                |___/

func TestFlagBuildArgs_Name(t *testing.T) {
	argsNames := FlagBuildArgs{
		Names: []string{"test"},
	}
	assert.Equal(t, "test", argsNames.Name())
	argsNoNames := FlagBuildArgs{
		Names: []string{},
	}
	assert.Equal(t, "<arg>", argsNoNames.Name())
}

//  _____     _           __     __    _
// |_   _|_ _| | _____  __\ \   / /_ _| |_   _  ___
//   | |/ _` | |/ / _ \/ __\ \ / / _` | | | | |/ _ \
//   | | (_| |   <  __/\__ \\ V / (_| | | |_| |  __/
//   |_|\__,_|_|\_\___||___/ \_/ \__,_|_|\__,_|\___|

func TestFlag_TakesValue(t *testing.T) {
	takerVal := intValue(12)
	taker := Flag{
		FlagBuildArgs: FlagBuildArgs{},
		Default:       &takerVal,
		Value:         &takerVal,
	}
	assert.True(t, taker.TakesValue())
	boolTakerVal := boolValue(true)
	boolTaker := Flag{
		FlagBuildArgs: FlagBuildArgs{},
		Default:       &boolTakerVal,
		Value:         &boolTakerVal,
	}
	assert.True(t, boolTaker.TakesValue())
	boolNonTakerVal := boolValue(false)
	boolNonTaker := Flag{
		FlagBuildArgs: FlagBuildArgs{},
		Default:       &boolNonTakerVal,
		Value:         &boolNonTakerVal,
	}
	assert.False(t, boolNonTaker.TakesValue())
}

//  _____ _                              _   _               _
// |  ___| | __ _  __ _   _ __ ___   ___| |_| |__   ___   __| |___
// | |_  | |/ _` |/ _` | | '_ ` _ \ / _ \ __| '_ \ / _ \ / _` / __|
// |  _| | | (_| | (_| | | | | | | |  __/ |_| | | | (_) | (_| \__ \
// |_|   |_|\__,_|\__, | |_| |_| |_|\___|\__|_| |_|\___/ \__,_|___/
//                |___/

// Checks X and PositionalX methods
func TestFlagSet_FlagMethods(t *testing.T) {
	test := func(
		conFlag func(*FlagSet, FlagBuildArgs, interface{}) interface{},
		conPos func(*FlagSet, FlagBuildArgs, interface{}) interface{},
		// this is necessary because there are no generics.
		// comparison between interfaces will be done in a separate functions that knows the types.
		eq func(interface{}, FlagValue) bool,
		vals ...interface{},
	) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			args := FlagBuildArgs{
				Names: []string{"a", "b"},
			}
			for _, val := range vals {
				fs := CreateFlagSet()
				p := conPos(fs, args, val)
				assert.Empty(t, fs.flags)
				if assert.Len(t, fs.positional, 1) {
					f := fs.positional[0]
					assert.EqualValues(t, p, f.Value)
					assert.True(t, eq(val, f.Value))
					assert.True(t, eq(val, f.Default))
				}

				fs = CreateFlagSet()
				p = conFlag(fs, args, val)
				assert.Empty(t, fs.positional)
				if assert.Len(t, fs.flags, 2) {
					if assert.Contains(t, fs.flags, "a") {
						f := fs.flags["a"]
						assert.EqualValues(t, p, f.Value)
						assert.True(t, eq(val, f.Value))
						assert.True(t, eq(val, f.Default))
					}
					if assert.Contains(t, fs.flags, "b") {
						f := fs.flags["b"]
						assert.EqualValues(t, p, f.Value)
						assert.True(t, eq(val, f.Value))
						assert.True(t, eq(val, f.Default))
					}
				}
			}
		}
	}

	t.Run("Bool", test(
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.Bool(args, def.(bool))
		},
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.PositionalBool(args, def.(bool))
		},
		func(a interface{}, b FlagValue) bool {
			return a.(bool) == bool(*b.(*boolValue))
		},
		true, false,
	))
	t.Run("Int", test(
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.Int(args, def.(int))
		},
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.PositionalInt(args, def.(int))
		},
		func(a interface{}, b FlagValue) bool {
			return a.(int) == int(*b.(*intValue))
		},
		123, -123,
	))
	t.Run("Uint", test(
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.Uint(args, def.(uint))
		},
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.PositionalUint(args, def.(uint))
		},
		func(a interface{}, b FlagValue) bool {
			return a.(uint) == uint(*b.(*uintValue))
		},
		uint(123), uint(4567890),
	))
	t.Run("String", test(
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.String(args, def.(string))
		},
		func(fs *FlagSet, args FlagBuildArgs, def interface{}) interface{} {
			return fs.PositionalString(args, def.(string))
		},
		func(a interface{}, b FlagValue) bool {
			return a.(string) == string(*b.(*stringValue))
		},
		"Hello", "World!", "",
	))
}

//            _     _ _____ _                ______           _ _   _                   _
//   __ _  __| | __| |  ___| | __ _  __ _   / /  _ \ ___  ___(_) |_(_) ___  _ __   __ _| |
//  / _` |/ _` |/ _` | |_  | |/ _` |/ _` | / /| |_) / _ \/ __| | __| |/ _ \| '_ \ / _` | |
// | (_| | (_| | (_| |  _| | | (_| | (_| |/ / |  __/ (_) \__ \ | |_| | (_) | | | | (_| | |
//  \__,_|\__,_|\__,_|_|   |_|\__,_|\__, /_/  |_|   \___/|___/_|\__|_|\___/|_| |_|\__,_|_|
//                                  |___/

func TestFlagSet_addFlag(t *testing.T) {
	val := intValue(123)
	fs := CreateFlagSet()
	flag := Flag{
		FlagBuildArgs: FlagBuildArgs{
			Names: []string{"test1", "test2"},
			Usage: "Hello World!",
		},
		Default: &val,
		Value:   &val,
	}
	expected := flag // copy original
	fs.addFlag(&flag)
	assert.Empty(t, fs.positional)
	if assert.Len(t, fs.flags, 2) {
		def := "123"
		expected.DefaultText = &def
		if assert.Contains(t, fs.flags, "test1") {
			assert.Equal(t, expected, *fs.flags["test1"])
		}
		if assert.Contains(t, fs.flags, "test2") {
			assert.Equal(t, expected, *fs.flags["test2"])
		}
	}

	fs = CreateFlagSet()
	def := "7B"
	flag = Flag{
		FlagBuildArgs: FlagBuildArgs{
			Names:       []string{"test1"},
			DefaultText: &def,
		},
		Default: &val,
		Value:   &val,
	}
	expected = flag // copy original
	fs.addFlag(&flag)
	assert.Empty(t, fs.positional)
	if assert.Len(t, fs.flags, 1) {
		if assert.Contains(t, fs.flags, "test1") {
			assert.Equal(t, expected, *fs.flags["test1"])
		}
	}

	fs = CreateFlagSet()
	flag = Flag{
		FlagBuildArgs: FlagBuildArgs{},
		Default:       &val,
		Value:         &val,
	}
	fs.addFlag(&flag)
	assert.Empty(t, fs.positional)
	assert.Empty(t, fs.flags)
}

func TestFlagSet_addPositional(t *testing.T) {
	val := intValue(123)
	fs := CreateFlagSet()
	flag := Flag{
		FlagBuildArgs: FlagBuildArgs{
			Names: []string{"test1", "test2"},
			Usage: "Hello World!",
		},
		Default: &val,
		Value:   &val,
	}
	expected := flag // copy original
	fs.addPositional(&flag)
	assert.Empty(t, fs.flags)
	if assert.Equal(t, 1, len(fs.positional)) {
		def := "123"
		expected.DefaultText = &def
		assert.Equal(t, flag, *fs.positional[0])
	}

	fs = CreateFlagSet()
	def := "7B"
	flag = Flag{
		FlagBuildArgs: FlagBuildArgs{
			DefaultText: &def,
		},
		Default: &val,
		Value:   &val,
	}
	expected = flag // copy original
	fs.addPositional(&flag)
	assert.Empty(t, fs.flags)
	if assert.Equal(t, 1, len(fs.positional)) {
		assert.Equal(t, expected, *fs.positional[0])
	}
}
