package main

import (
	"fmt"

	"github.com/getoptimum/optimum-common/pkg/version"
)

func main() {
	fmt.Printf("Version: %s\n", version.GetVersion())
	fmt.Printf("Commit:  %s\n\n", version.GetCommitHash())

	inputs := []string{
		"v1.2.3",
		"v1.2.3-rc1",
		"v1.2.3-0.20240102112233-deadbeef",
		"v0.0.0-20240102112233-deadbeef",
		"v1.2.3+incompatible",
	}

	fmt.Println("Derived versions:")
	for _, in := range inputs {
		fmt.Printf("  %-40s -> %s\n", in, version.DeriveVersion(in))
	}
}
