package crawl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleAgentVersionParsing() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`go-ipfs/0.9.0/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`/go-ipfs/0.5.0-dev/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`no-match`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"go-ipfs/0.9.0/ce693d", "0.9.0", "", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
	// []string{"/go-ipfs/0.5.0-dev/ce693d", "0.5.0", "dev", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
	// []string(nil)
	// []string{"", "core", "prerelease", "commit"}
}

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
			s := &Scheduler{errors: tt.errors}
			assert.Equal(t, tt.want, s.TotalErrors())
		})
	}
}
