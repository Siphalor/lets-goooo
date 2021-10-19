package util

import (
	"errors"
	"fmt"
	"os"
)

func FileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists %w", err)
	}
	return !stat.IsDir(), nil
}
