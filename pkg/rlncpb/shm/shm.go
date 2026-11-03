package shm

import "fmt"

func (e *OpError) Error() string {
	if e == nil {
		return ""
	}

	return fmt.Sprintf("OpError{code: %d, err: %s}", e.GetCode(), e.GetErr())
}
