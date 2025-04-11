// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"

	kitconfig "github.com/fsyyft-go/kit/config"
)

func main() {
	fmt.Printf("Version: %s\n", kitconfig.CurrentVersion.Version())
	fmt.Printf("Git Version: %s\n", kitconfig.CurrentVersion.GitVersion())
	fmt.Printf("Build Time: %s\n", kitconfig.CurrentVersion.BuildTimeString())
	fmt.Printf("Library Directory: %s\n", kitconfig.CurrentVersion.BuildLibraryDirectory())
	fmt.Printf("Working Directory: %s\n", kitconfig.CurrentVersion.BuildWorkingDirectory())
	fmt.Printf("GOPATH Directory: %s\n", kitconfig.CurrentVersion.BuildGopathDirectory())
	fmt.Printf("GOROOT Directory: %s\n", kitconfig.CurrentVersion.BuildGorootDirectory())
	fmt.Printf("Debug Mode: %v\n", kitconfig.CurrentVersion.Debug())
}
