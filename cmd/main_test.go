package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindModules(t *testing.T) {
	modules := Modules{
		Modules: []Module{
			{Key: "module1", Source: "github.com/example/module1", Version: "v1.0.0", Dir: "module1"},
			{Key: "module2", Source: "github.com/example/module2", Version: "v2.0.0", Dir: "module2"},
		},
	}

	expected := []foundModule{{ModuleName: "github.com/example/\x1b[31mmodule1\x1b[0m", Version: "v1.0.0", ModuleLocalName: "'module1'"}}
	got := formatModules(modules, "module1")

	assert.Equal(t, expected, got)
}
