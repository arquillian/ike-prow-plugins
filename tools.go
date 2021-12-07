//go:build tools
// +build tools

package tools

// nolint
import (
	_ "github.com/go-bindata/go-bindata/v3"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "github.com/onsi/ginkgo/v2/ginkgo/generators"
	_ "golang.org/x/tools/cmd/goimports"
)
