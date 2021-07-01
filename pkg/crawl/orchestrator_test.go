package crawl

import (
	"fmt"
)

func ExampleAgentVersionParsing1() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`go-ipfs/0.9.0/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"go-ipfs/0.9.0/ce693d", "0.9.0", "", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
}

func ExampleAgentVersionParsing2() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`/go-ipfs/0.5.0-dev/ce693d`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"/go-ipfs/0.5.0-dev/ce693d", "0.5.0", "dev", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
}

func ExampleAgentVersionParsing3() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`/go-ipfs/0.9.0/`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"/go-ipfs/0.5.0-dev/ce693d", "0.9.0", "", ""}
	// []string{"", "core", "prerelease", "commit"}
}

func ExampleAgentVersionParsing4() {
	fmt.Printf("%#v\n", agentVersionRegex.FindStringSubmatch(`no-match`))
	fmt.Printf("%#v\n", agentVersionRegex.SubexpNames())
	// Output:
	// []string{"/go-ipfs/0.5.0-dev/ce693d", "0.5.0", "dev", "ce693d"}
	// []string{"", "core", "prerelease", "commit"}
}
