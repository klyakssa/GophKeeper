package luhn_test

import (
	"gophkeeper/pkg/luhn"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	tests := []struct {
		name         string
		numberString string
		want         bool
	}{
		{
			name:         "valid card number",
			numberString: "4539 1488 0343 6467",
			want:         true,
		},
		{
			name:         "invalid card number",
			numberString: "4539 1488 0343 6468",
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luhn.Valid(tt.numberString)
			assert.Equal(t, tt.want, got)
		})
	}
}
