package vyos

import "github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"

// CommandError pairs an SSH failure with the VyOS command output needed to diagnose it.
type CommandError struct {
	Result sshclient.Result
	Err    error
}

func (e *CommandError) Error() string { return "VyOS command failed" }
func (e *CommandError) Unwrap() error { return e.Err }
