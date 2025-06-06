package florist

import (
	"errors"
	"fmt"
	"strings"
)

// JoinErrors returns an error containing the joined error messages of 'errs'.
// The error is formatted as the concatenation of the strings obtained by calling the
// Error() method of each element of errs, with a semicolon (;) between each string.
// If all elements of 'errs' are nil, JoinErrors returns nil.
// Does not preserve the error types nor provides Unwrap.
func JoinErrors(errs ...error) error {
	errors.Join()
	var msgs []string
	for _, err := range errs {
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}
	if len(msgs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}
