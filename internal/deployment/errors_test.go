package deployment_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
)

func TestStageErrorPreservesCauseWithoutPrintingDiagnostics(t *testing.T) {
	cause := errors.New("native rejection: password sensitive-value")
	err := fmt.Errorf("deploy file: %w", &deployment.StageError{
		Stage: deployment.StageCommit,
		Err:   cause,
	})
	if strings.Contains(err.Error(), "sensitive-value") {
		t.Fatal("wrapped deployment error exposed device diagnostics")
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped deployment error lost the original cause")
	}
	var stage *deployment.StageError
	if !errors.As(err, &stage) || stage.Stage != deployment.StageCommit {
		t.Fatal("wrapped deployment error lost the failing stage")
	}
}
