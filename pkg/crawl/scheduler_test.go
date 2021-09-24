package crawl

import (
	"fmt"
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
