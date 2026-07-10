package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructEnabled())

func Validate(s any) error {
	if err := v.Struct(s); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		ves := errs
		msgs := make([]string, 0, len(ves))
		for _, fe := range ves {
			msgs = append(msgs, fmt.Sprintf("%s: failed %q", fe.Namespace(), fe.Tag()))
		}
		return fmt.Errorf("config validation:\n  - %s", strings.Join(msgs, "\n  - "))
	}
	return nil
}
