package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type deployOptions struct {
	File, Host, User, Identity, KnownHosts         string
	ConfirmTimeout, ConnectTimeout, CommandTimeout time.Duration
	HostKeyAlgorithms                              string
	Save                                           bool
}

func parseDeployOptions(args []string, output io.Writer) (deployOptions, error) {
	flags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {
		flags.SetOutput(output)
		flags.PrintDefaults()
	}
	host := flags.String("host", "", "required: device SSH address (host:port)")
	user := flags.String("user", "", "required: SSH username")
	identity := flags.String("identity", "", "required: unencrypted SSH private key file")
	hosts := flags.String("known-hosts", "", "required: trusted SSH host keys file")
	timeout := flags.Duration("confirm-timeout", 5*time.Minute, "verification window (whole minutes)")
	connect := flags.Duration("connect-timeout", 30*time.Second, "SSH connection timeout")
	command := flags.Duration("command-timeout", 30*time.Second, "device operation timeout")
	algorithms := flags.String("host-key-algorithms", "", "comma-separated SSH host key algorithms (match the trusted device key)")
	save := flags.Bool("save", false, "persist configuration after acceptance")
	rest := args
	var file string
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		file = rest[0]
		rest = rest[1:]
	}
	if err := flags.Parse(rest); err != nil {
		return deployOptions{}, err
	}
	if file == "" && flags.NArg() == 1 {
		file = flags.Arg(0)
	} else if flags.NArg() != 0 {
		return deployOptions{}, errors.New("provide exactly one configuration file")
	}
	options := deployOptions{
		File: file, Host: *host, User: *user, Identity: *identity, KnownHosts: *hosts,
		ConfirmTimeout: *timeout, ConnectTimeout: *connect, CommandTimeout: *command,
		HostKeyAlgorithms: *algorithms, Save: *save,
	}
	return options, options.validate()
}

func (o deployOptions) validate() error {
	var issues []error
	for _, field := range []struct{ name, value string }{
		{"configuration file", o.File}, {"--host", o.Host}, {"--user", o.User},
		{"--identity", o.Identity}, {"--known-hosts", o.KnownHosts},
	} {
		if strings.TrimSpace(field.value) == "" {
			issues = append(issues, fmt.Errorf("%s is required", field.name))
		}
	}
	if strings.TrimSpace(o.Host) != "" {
		hostname, port, err := net.SplitHostPort(o.Host)
		switch {
		case err != nil:
			issues = append(issues, errors.New("--host must use host:port (enclose IPv6 addresses in brackets)"))
		case strings.TrimSpace(hostname) == "":
			issues = append(issues, errors.New("--host must include a hostname or IP address before the port"))
		default:
			if strings.ContainsAny(hostname, " \t\r\n") {
				issues = append(issues, errors.New("--host hostname or IP address must not contain whitespace"))
			}
			portNumber, err := strconv.Atoi(port)
			if err != nil || portNumber < 1 || portNumber > 65535 {
				issues = append(issues, errors.New("--host port must be an integer between 1 and 65535"))
			}
		}
	}
	for _, field := range []struct {
		name  string
		value time.Duration
	}{
		{"--confirm-timeout", o.ConfirmTimeout}, {"--connect-timeout", o.ConnectTimeout}, {"--command-timeout", o.CommandTimeout},
	} {
		if field.value <= 0 {
			issues = append(issues, fmt.Errorf("%s must be greater than zero", field.name))
		}
	}
	if o.ConfirmTimeout > 0 && o.ConfirmTimeout%time.Minute != 0 {
		issues = append(issues, errors.New("--confirm-timeout must be a whole number of minutes (for example, 5m)"))
	}
	if o.HostKeyAlgorithms != "" {
		supported := ssh.SupportedAlgorithms().HostKeys
		for _, algorithm := range strings.Split(o.HostKeyAlgorithms, ",") {
			algorithm = strings.TrimSpace(algorithm)
			if !slices.Contains(supported, algorithm) {
				issues = append(issues, fmt.Errorf("--host-key-algorithms: %q is unsupported; choose from %s", algorithm, strings.Join(supported, ", ")))
			}
		}
	}
	return errors.Join(issues...)
}

func loadDeployment(options deployOptions) (deployment.Request, sshclient.Config, error) {
	content, err := os.ReadFile(options.File)
	if err != nil {
		return deployment.Request{}, sshclient.Config{}, fmt.Errorf("configuration file: %w", err)
	}
	if len(content) == 0 {
		return deployment.Request{}, sshclient.Config{}, errors.New("configuration file is empty")
	}
	key, err := os.ReadFile(options.Identity)
	if err != nil {
		return deployment.Request{}, sshclient.Config{}, fmt.Errorf("--identity: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		var encrypted *ssh.PassphraseMissingError
		if errors.As(err, &encrypted) {
			return deployment.Request{}, sshclient.Config{}, errors.New("--identity is encrypted; passphrase-protected keys are not supported")
		}
		return deployment.Request{}, sshclient.Config{}, errors.New("--identity does not contain a supported SSH private key")
	}
	hostKey, err := knownhosts.New(options.KnownHosts)
	if err != nil {
		var fileError *os.PathError
		if errors.As(err, &fileError) {
			return deployment.Request{}, sshclient.Config{}, fmt.Errorf("--known-hosts: %w", fileError)
		}
		// Parser errors can include file contents, so only expose the format failure.
		return deployment.Request{}, sshclient.Config{}, errors.New("--known-hosts contains an invalid host key entry")
	}
	var hostKeyAlgorithms []string
	if options.HostKeyAlgorithms != "" {
		for _, algorithm := range strings.Split(options.HostKeyAlgorithms, ",") {
			algorithm = strings.TrimSpace(algorithm)
			hostKeyAlgorithms = append(hostKeyAlgorithms, algorithm)
		}
	}
	request := deployment.Request{
		Artifact: deployment.Artifact{Content: content}, ConfirmTimeout: options.ConfirmTimeout, SaveAfterConfirm: options.Save,
	}
	connection := sshclient.Config{
		Address: options.Host, User: options.User, Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKey: hostKey, HostKeyAlgorithms: hostKeyAlgorithms,
		ConnectTimeout: options.ConnectTimeout, CommandTimeout: options.CommandTimeout,
	}
	return request, connection, nil
}
