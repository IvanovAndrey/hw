package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var (
	ErrInputNoStruct          = errors.New("input must be a struct or pointer to struct")
	ErrRegexpNotMatch         = errors.New("must match regexp")
	ErrInvalidLenValue        = errors.New("invalid length value")
	ErrInvalidRegexpPattern   = errors.New("invalid regexp")
	ErrUnknownStringValidator = errors.New("unknown validator for string")
	ErrValueNotAllowed        = errors.New("value is not in the allowed list")
	ErrInvalidMinValue        = errors.New("invalid min value")
	ErrInvalidMaxValue        = errors.New("invalid max value")
	ErrInvalidInValue         = errors.New("invalid in value")
	ErrUnknownIntValidator    = errors.New("unknown validator for int")
	ErrUnknownValidatorType   = errors.New("unknown validator type")
	ErrUnsupportedSliceElem   = errors.New("unsupported slice element type")
)

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for _, err := range v {
		sb.WriteString(fmt.Sprintf("%s: %v; ", err.Field, err.Err))
	}
	return sb.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ErrInputNoStruct
	}

	var validationErrors ValidationErrors

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := strings.Split(validateTag, "|")
		for _, rule := range rules {
			parts := strings.SplitN(rule, ":", 2)
			if len(parts) != 2 {
				continue
			}

			validator := parts[0]
			value := parts[1]

			var err error
			var errSlice []error

			//nolint:exhaustive
			switch fieldValue.Kind() {
			case reflect.String:
				err = validateString(fieldValue.String(), validator, value)
			case reflect.Int:
				err = validateInt(fieldValue.Int(), validator, value)
			case reflect.Slice:
				errSlice = validateSlice(fieldValue, validator, value)
			default:
				err = fmt.Errorf("%w: %v", ErrUnknownValidatorType, fieldValue.Kind())
			}

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{
					Field: field.Name,
					Err:   fmt.Errorf("%s(%s): %w", validator, value, err),
				})
			}
			if len(errSlice) > 0 {
				for _, err := range errSlice {
					validationErrors = append(validationErrors, ValidationError{
						Field: field.Name,
						Err:   fmt.Errorf("%s(%s): %w", validator, value, err),
					})
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

func validateString(s, validator, value string) error {
	switch validator {
	case "len":
		expectedLen, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidLenValue, err)
		}
		if len(s) != expectedLen {
			return fmt.Errorf("length must be %d", expectedLen)
		}
	case "regexp":
		matched, err := regexp.MatchString(value, s)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRegexpPattern, err)
		}
		if !matched {
			return fmt.Errorf("%w %s", ErrRegexpNotMatch, value)
		}
	case "in":
		allowedValues := strings.Split(value, ",")
		found := false
		for _, v := range allowedValues {
			if s == v {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: %v", ErrValueNotAllowed, allowedValues)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownStringValidator, validator)
	}
	return nil
}

func validateInt(i int64, validator, value string) error {
	switch validator {
	case "min":
		minVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidMinValue, err)
		}
		if i < minVal {
			return fmt.Errorf("must be >= %d", minVal)
		}
	case "max":
		maxVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidMaxValue, err)
		}
		if i > maxVal {
			return fmt.Errorf("must be <= %d", maxVal)
		}
	case "in":
		allowedValues := strings.Split(value, ",")
		found := false
		for _, v := range allowedValues {
			num, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidInValue, err)
			}
			if i == num {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: %v", ErrValueNotAllowed, allowedValues)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnknownIntValidator, validator)
	}
	return nil
}

func validateSlice(slice reflect.Value, validator, value string) []error {
	var res []error
	for i := 0; i < slice.Len(); i++ {
		element := slice.Index(i)
		var err error
		//nolint:exhaustive
		switch element.Kind() {
		case reflect.String:
			err = validateString(element.String(), validator, value)
		case reflect.Int:
			err = validateInt(element.Int(), validator, value)
		default:
			res = append(res, fmt.Errorf("%w: %s", ErrUnsupportedSliceElem, element.Kind()))
		}

		if err != nil {
			res = append(res, fmt.Errorf("element %d: %w", i, err))
		}
	}
	return res
}
