package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	InvalidStruct struct {
		Value float64 `validate:"min:10"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "Valid User",
				Age:    25,
				Email:  "valid@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid user (multiple errors)",
			in: User{
				ID:     "short",
				Name:   "Invalid User",
				Age:    15,
				Email:  "invalid-email",
				Role:   "unknown",
				Phones: []string{"123", "123456789012"},
			},
			expectedErr: ValidationErrors{
				{Field: "ID", Err: errors.New("len(36): length must be 36")},
				{Field: "Age", Err: errors.New("min(18): must be >= 18")},
				{Field: "Email", Err: errors.New("regexp(^\\w+@\\w+\\.\\w+$): must match regexp ^\\w+@\\w+\\.\\w+$")},
				{Field: "Role", Err: errors.New("in(admin,stuff): value is not in the allowed list: [admin stuff]")},
				{Field: "Phones", Err: errors.New("len(11): element 0: length must be 11")},
				{Field: "Phones", Err: errors.New("len(11): element 1: length must be 11")},
			},
		},
		{
			name: "valid app",
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			name: "invalid app",
			in: App{
				Version: "toolong",
			},
			expectedErr: ValidationErrors{
				{Field: "Version", Err: errors.New("len(5): length must be 5")},
			},
		},
		{
			name: "valid token (no validation)",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			name: "invalid response",
			in: Response{
				Code: 400,
				Body: "Bad Request",
			},
			expectedErr: ValidationErrors{
				{Field: "Code", Err: errors.New("in(200,404,500): value is not in the allowed list: [200 404 500]")},
			},
		},
		{
			name:        "invalid struct type",
			in:          "not a struct",
			expectedErr: errors.New("input must be a struct or pointer to struct"),
		},
		{
			name: "unsupported field type",
			in:   InvalidStruct{Value: 5.5},
			expectedErr: ValidationErrors{
				{Field: "Value", Err: errors.New("min(10): unknown validator type: float64")},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			var expectedValidationErrors ValidationErrors
			if errors.As(tt.expectedErr, &expectedValidationErrors) {
				var actual ValidationErrors
				ok := errors.As(err, &actual)
				require.True(t, ok, "expected ValidationErrors")

				require.Equal(t, len(expectedValidationErrors), len(actual), "number of errors mismatch")

				for i := range expectedValidationErrors {
					require.Equal(t, expectedValidationErrors[i].Field, actual[i].Field, "field name mismatch")
					require.EqualError(t, actual[i].Err, expectedValidationErrors[i].Err.Error(), "error message mismatch")
				}
				return
			}

			require.EqualError(t, err, tt.expectedErr.Error())

			_ = tt
		})
	}
}
