//go:build generate
// +build generate

// Package graph code generation dependencies
// This file ensures dev tools remain in go.mod for code generation workflows
package graph

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
)
