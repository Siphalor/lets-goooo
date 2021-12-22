// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package argp

import "strconv"

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
