package validation

import "fmt"

func StringSize(name string, s string, size int) error {
	if len(s) > size {
		return fmt.Errorf("'%s' size is greater than %d", name, size)
	}
	return nil
}

func StringNotEmpty(name string, s string) error {
	if len(s) == 0 {
		return fmt.Errorf("'%s' cannot be empty", name)
	}
	return nil
}
