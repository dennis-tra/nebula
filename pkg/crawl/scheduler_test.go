package crawl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_TotalErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]int
		want   int
	}{
		{
			name:   "Nil map",
			errors: nil,
			want:   0,
		},
		{
			name:   "Empty map",
			errors: map[string]int{},
			want:   0,
		},
		{
			name: "One entry",
			errors: map[string]int{
				"some-error": 1,
			},
			want: 1,
		},
		{
			name: "Multiple entry",
			errors: map[string]int{
				"some-error-1": 5,
				"some-error-2": 4,
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{connErrs: tt.errors}
			assert.Equal(t, tt.want, s.TotalErrors())
		})
	}
}
