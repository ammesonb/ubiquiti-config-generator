package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/logging"
	"github.com/ammesonb/ubiquiti-config-generator/internal/vyos"
)

// Streams supplies the terminal used to review and accept a deployment.
type Streams struct {
	Input       io.Reader
	Output      io.Writer
	Error       io.Writer
	Interactive bool
}

// Run deploys a complete configuration, returning a process exit status.
func Run(ctx context.Context, args []string, streams Streams) int {
	errorsLog := logging.New(streams.Error)
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		_, _ = fmt.Fprintln(streams.Output, "Usage: ubiquiti-config-generator deploy FILE --host HOST:PORT --user USER --identity KEY --known-hosts FILE [--confirm-timeout 5m] [--save]\n\nVerify connectivity and intended settings on the device before accepting the changes.\nWithout acceptance, the device's rollback timer remains enabled.\n--save persists the configuration only after acceptance. Interactive input is required.")
		return 0
	}
	if args[0] != "deploy" {
		_ = errorsLog.Error("expected deploy subcommand")
		return 2
	}
	options, err := parseDeployOptions(args[1:], streams.Error)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		_ = errorsLog.Error(err.Error())
		return 2
	}
	if !streams.Interactive {
		_ = errorsLog.Error("deployment requires an interactive terminal for confirmation")
		return 2
	}
	request, connection, err := loadDeployment(options)
	if err != nil {
		_ = errorsLog.Error(err.Error())
		return 2
	}
	runner := deployment.Runner{Dialer: vyos.Dialer{SSH: connection}}
	outcome, err := runner.Deploy(ctx, request, approval(streams, options.Save))
	return report(streams, outcome, err)
}

func approval(streams Streams, save bool) deployment.Approval {
	return func(ctx context.Context) <-chan deployment.Decision {
		decisions := make(chan deployment.Decision, 1)
		policy := "**IMPORTANT**: Accepting changes will disable rollback, requires restart to revert."
		if save {
			policy = "**WARNING**: Accepting changes will disable rollback and automatically persist the configuration."
		}
		log := logging.New(streams.Output)
		err := log.Info("Changes are active. Check device connectivity and intended settings now.")
		if err == nil {
			if save {
				err = log.Warn(policy)
			} else {
				err = log.Info(policy)
			}
		}
		if err == nil {
			_, err = fmt.Fprintln(streams.Output, "Confirm these changes? [y/N]")
		}

		if err != nil {
			decisions <- deployment.Decision{Err: err}
			return decisions
		}
		// Stdin reads cannot be canceled portably. A buffered reply lets this one-shot
		// command exit on timeout without waiting for another line of input.
		go func() {
			line, err := bufio.NewReader(streams.Input).ReadString('\n')
			if errors.Is(err, io.EOF) {
				err = nil
				line = ""
			}
			if ctx.Err() != nil {
				return
			}
			answer := strings.ToLower(strings.TrimSpace(line))
			decisions <- deployment.Decision{Accepted: answer == "y" || answer == "yes", Err: err}
		}()
		return decisions
	}
}

func report(streams Streams, outcome deployment.Outcome, err error) int {
	log := logging.New(streams.Output)
	errorsLog := logging.New(streams.Error)
	if outcome.Saved {
		_ = log.Info("Configuration confirmed and saved across restart.")
	} else if outcome.Confirmed && err == nil {
		_ = log.Info("Configuration confirmed without saving across restart.")
	} else if outcome.Confirmed {
		_ = log.Warn("Configuration confirmed, but the requested save failed - check for inconsistencies!")
	}
	if err == nil {
		return 0
	}
	_ = errorsLog.Error(err.Error())
	switch {
	case errors.Is(err, deployment.ErrDeclined):
		_ = errorsLog.Warn("Changes were not accepted. Automatic rollback remains enabled.")
	case errors.Is(err, deployment.ErrConfirmationEnded):
		_ = errorsLog.Warn("Confirmation is no longer pending. Inspect the device to determine its current configuration.")
	}
	if !outcome.Confirmed {
		_ = errorsLog.Warn("This command did not confirm the changes. Inspect device state before retrying; rollback may still be pending.")
	}
	return 1
}
