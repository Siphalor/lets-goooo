// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package argp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//  _                 ___     __    _
// | |__   ___   ___ | \ \   / /_ _| |_   _  ___
// | '_ \ / _ \ / _ \| |\ \ / / _` | | | | |/ _ \
// | |_) | (_) | (_) | | \ V / (_| | | |_| |  __/
// |_.__/ \___/ \___/|_|  \_/ \__,_|_|\__,_|\___|

func TestBoolValue_String(t *testing.T) {
	var val boolValue
	val = true
	assert.Equal(t, "true", val.String())
	val = false
	assert.Equal(t, "false", val.String())
}

func TestBoolValue_FromString(t *testing.T) {
	var val boolValue
	val = true
	if assert.NoError(t, val.FromString("false")) {
		assert.Equal(t, boolValue(false), val)
	}
	val = true
	if assert.NoError(t, val.FromString("0")) {
		assert.Equal(t, boolValue(false), val)
	}
	val = false
	if assert.NoError(t, val.FromString("true")) {
		assert.Equal(t, boolValue(true), val)
	}
	val = false
	if assert.NoError(t, val.FromString("1")) {
		assert.Equal(t, boolValue(true), val)
	}
	assert.Error(t, val.FromString("not a bool"))
}

//   __   __  _       _ __     __    _
//  / /   \ \(_)_ __ | |\ \   / /_ _| |_   _  ___
// | | | | | | | '_ \| __\ \ / / _` | | | | |/ _ \
// | | |_| | | | | | | |_ \ V / (_| | | |_| |  __/
// | |\__,_| |_|_| |_|\__| \_/ \__,_|_|\__,_|\___|
//  \_\   /_/

func TestUIntValue_String(t *testing.T) {
	test := func(f func(uint64) FlagValue) func(*testing.T) {
		return func(t *testing.T) {
			var val FlagValue
			val = f(123)
			assert.Equal(t, "123", val.String())
			val = f(2147483647)
			assert.Equal(t, "2147483647", val.String())
		}
	}

	t.Run("uint", test(func(v uint64) FlagValue {
		tmp := uintValue(v)
		return &tmp
	}))
	t.Run("int", test(func(v uint64) FlagValue {
		tmp := intValue(v)
		return &tmp
	}))
}

func TestUIntValue_FromString(t *testing.T) {
	test := func(con func(uint) FlagValue, parser func(text string) (FlagValue, error)) func(*testing.T) {
		return func(t *testing.T) {
			var val FlagValue
			var err error
			val, err = parser("123")
			assert.NoError(t, err)
			assert.Equal(t, con(123), val)
			val, err = parser("2147483648")
			assert.NoError(t, err)
			assert.Equal(t, con(2147483648), val)
			val, err = parser("0b101") // binary
			assert.NoError(t, err)
			assert.Equal(t, con(5), val)
			val, err = parser("0101") // octal
			assert.NoError(t, err)
			assert.Equal(t, con(65), val)
			val, err = parser("0x101") // hexadecimal
			assert.NoError(t, err)
			assert.Equal(t, con(257), val)
			_, err = parser("-9223372036854775809")
			assert.Error(t, err)
			_, err = parser("not a number!")
			assert.Error(t, err)
			_, err = parser("infinity")
			assert.Error(t, err)
		}
	}

	t.Run("uint", test(func(v uint) FlagValue {
		tmp := uintValue(v)
		return &tmp
	}, func(text string) (FlagValue, error) {
		tmp := uintValue(0)
		err := tmp.FromString(text)
		return &tmp, err
	}))
	t.Run("int", test(func(v uint) FlagValue {
		tmp := intValue(v)
		return &tmp
	}, func(text string) (FlagValue, error) {
		tmp := intValue(0)
		err := tmp.FromString(text)
		return &tmp, err
	}))
}

//      _        _           __     __    _
//  ___| |_ _ __(_)_ __   __ \ \   / /_ _| |_   _  ___
// / __| __| '__| | '_ \ / _` \ \ / / _` | | | | |/ _ \
// \__ \ |_| |  | | | | | (_| |\ V / (_| | | |_| |  __/
// |___/\__|_|  |_|_| |_|\__, | \_/ \__,_|_|\__,_|\___|
//                       |___/

func TestStringValue_String(t *testing.T) {
	val := stringValue("test")
	assert.Equal(t, "test", val.String())
}

func TestStringValue_FromString(t *testing.T) {
	val := stringValue("")

	simple := stringValue("simple")
	assert.NoError(t, val.FromString("simple"))
	assert.Equal(t, simple, val)
	advanced := stringValue("\"advanced\"")
	assert.NoError(t, val.FromString("\"advanced\""))
	assert.Equal(t, advanced, val)
}
