package util

import (
	"fmt"
	"os"
)

func FileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		if os.IsExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to check if file exists %w", err)
	}
	return !stat.IsDir(), nil
}
