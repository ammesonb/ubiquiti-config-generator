//go:build integration

package integration

import (
	"context"
	"flag"
	"path/filepath"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/testlab"
)

var dockerLab = flag.Bool("docker-lab", false, "provision a disposable local Docker VyOS lab")

func TestRouterSmoke(t *testing.T) {
	var cfg testlab.Config
	if *dockerLab {
		cfg = startDockerLab(t)
	} else {
		var err error
		cfg, err = testlab.Load(filepath.Join("..", ".env.integration"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := testlab.Smoke(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	t.Log("SSH host identity verified; active configuration has the expected hostname")
}
