package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructEnabled())

func Validate(s any) error {
	if err := v.Struct(s); err != nil {
		var ves validator.ValidationErrors
		if errs, ok := err.(validator.ValidationErrors); ok {
			ves = errs
		} else {
			return err
		}
		msgs := make([]string, 0, len(ves))
		for _, fe := range ves {
			msgs = append(msgs, fmt.Sprintf("%s: failed %q", fe.Namespace(), fe.Tag()))
		}
		return fmt.Errorf("config validation:\n  - %s", strings.Join(msgs, "\n  - "))
	}
	return nil
}
