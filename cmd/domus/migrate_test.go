package main

import (
	"context"
	"testing"
)

func TestRootCmd_SubcommandsAndMigrateCLI(t *testing.T) {
	root := newRootCmd()
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	if !names["serve"] || !names["migrate"] {
		t.Fatalf("expected serve and migrate subcommands, got %v", names)
	}

	t.Setenv("APP_DB_URL", "file:cli_migrate?mode=memory&cache=shared")

	root.SetArgs([]string{"migrate"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("domus migrate failed: %v", err)
	}
}
