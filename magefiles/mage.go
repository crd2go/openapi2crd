//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

// Default target to run when just "mage" is invoked
var Default = Build

// Environment variables with fallbacks
var (
	binaryDir  = getEnv("BINARY_DIR", "bin")
	binaryName = "openapi2crd"
	binaryPath = filepath.Join(binaryDir, binaryName)
	crdFile    = getEnv("CRD_FILE", "crds.yaml")
)

// Build compiles the openapi2crd binary if source files have changed.
func Build() error {
	// Replicates the Makefile's "build: clean $(BINARY_PATH)"
	mg.Deps(Clean)

	// Fetch all go files excluding vendor to mimic $(GO_FILES)
	goFiles, err := getGoFiles()
	if err != nil {
		return err
	}

	// Mage's built-in helper to check if sources are newer than the binary
	changed, err := target.Path(binaryPath, goFiles...)
	if err != nil {
		return err
	}

	if changed {
		fmt.Printf("==> Building %s...\n", binaryPath)
		if err := os.MkdirAll(binaryDir, 0755); err != nil {
			return err
		}
		return sh.Run("go", "build", "-o", binaryPath, "main.go")
	}

	return nil
}

// Crds generates CRDs from the config file using the compiled binary.
func Crds() error {
	mg.Deps(Build)
	fmt.Println("==> Generating CRDs...")
	return sh.Run(binaryPath, "--config", "config.yaml", "--output", crdFile)
}

// CrdsForce generates CRDs directly via "go run" with the force flag.
func CrdsForce() error {
	fmt.Println("==> Generating CRDs...")
	return sh.Run("go", "run", "main.go", "--config", "config.yaml", "--force", "--output", crdFile)
}

// Fmt formats all Go code using gci.
func Fmt() error {
	fmt.Println("==> Formatting code...")
	return sh.Run("go", "tool", "gci", "write", "-s", "standard", "-s", "default", "-s", "localmodule", ".")
}

// UnitTest runs unit tests with race detection and coverage flags.
func UnitTest() error {
	fmt.Println("==> Running unit tests...")

	args := []string{"test", "-race", "-cover"}

	// Append GO_TEST_FLAGS if they exist in the environment
	if testFlags := os.Getenv("GO_TEST_FLAGS"); testFlags != "" {
		args = append(args, strings.Fields(testFlags)...)
	}

	// Replicates $(PACKAGES) via standard Go wildcard
	args = append(args, "./...")

	return sh.Run("go", args...)
}

// GenMock generates mocks for interfaces using mockery.
func GenMock() error {
	fmt.Println("==> Generating mocks...")
	return sh.Run("go", "tool", "mockery", "--config", ".mockery.yaml")
}

// All runs gen-mock, fmt, unit-test, and build in strict sequential order.
func All() {
	// mg.SerialDeps forces them to run one after another, matching Makefile behavior
	mg.SerialDeps(GenMock, Fmt, UnitTest, Build)
}

// Clean removes built artifacts and generated CRD files.
func Clean() error {
	fmt.Println("==> Cleaning...")
	_ = os.Remove(binaryPath)
	_ = os.Remove(crdFile)
	return nil
}

// Ci runs the standard suite of CI checks.
func Ci() {
	mg.Deps(All)
}

// --- Helpers ---

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Replicates the `find . -name '*.go' -not -path './vendor/*'` logic
func getGoFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor directory completely
		if info.IsDir() && (info.Name() == "vendor" || path == "vendor") {
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
