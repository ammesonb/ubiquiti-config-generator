package vyos

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// TestAddOption uses basic configuration stubs to verify variables are set properly
func TestAddOption(t *testing.T) {
	n := &Node{
		Name:        "TestNode",
		IsTag:       false,
		Multi:       false,
		Children:    map[string]*Node{},
		Constraints: []NodeConstraint{},
	}
	helpValues := &[]string{}
	allowed := ""
	expression := ""

	t.Run("tag", func(t *testing.T) {
		addOption("tag", "tag:", helpValues, &allowed, &expression, n)
		if !n.IsTag {
			t.Errorf("Tag flag not set on node")
		}
	})
	t.Run("multi", func(t *testing.T) {
		addOption("multi", "multi:", helpValues, &allowed, &expression, n)
		if !n.Multi {
			t.Errorf("Multi flag not set on node")
		}
	})
	t.Run("type", func(t *testing.T) {
		addOption("type", "type: txt", helpValues, &allowed, &expression, n)
		if n.Type != "txt" {
			t.Errorf("Expected node type to be txt, got '%s'", n.Type)
		}
	})
	t.Run("help", func(t *testing.T) {
		addOption("help", "help: some description", helpValues, &allowed, &expression, n)
		if n.Help != "some description" {
			t.Errorf("Got unexpected node description: '%s'", n.Type)
		}
	})
	t.Run("val_help", func(t *testing.T) {
		addOption("val_help", "val_help: u32", helpValues, &allowed, &expression, n)
		if len(*helpValues) == 0 {
			t.Errorf("No value added to help descriptions")
		} else if len(*helpValues) > 1 {
			t.Errorf("Too many values added to help descriptions: %#v", *helpValues)
		} else if (*helpValues)[0] != "u32" {
			t.Errorf("Got unexpected help value: %s", (*helpValues)[0])
		}
	})
	t.Run("allowed", func(t *testing.T) {
		addOption("allowed", "allowed: cli-shell-api", helpValues, &allowed, &expression, n)
		if allowed != "cli-shell-api" {
			t.Errorf("Got unexpected allowed value: '%s'", allowed)
		}
	})
	t.Run("syntax", func(t *testing.T) {
		addOption("syntax", "syntax:expression: pattern $VAR(@)", helpValues, &allowed, &expression, n)
		if expression != "pattern $VAR(@)" {
			t.Errorf("Got unexpected expression value: '%s'", expression)
		}
	})
}

func getBlank() string {
	if rand.Intn(2) == 1 {
		return "\n"
	}

	return ""
}

func constructDefinition(
	ntype *string,
	help *string,
	isTag bool,
	isMulti bool,
	helpValues []string,
	allowed *string,
	syntax *string,
) string {
	value := ""
	if ntype != nil {
		value += fmt.Sprintf("type: %s\n%s", *ntype, getBlank())
	}
	if help != nil {
		value += fmt.Sprintf("help: %s\n%s", *help, getBlank())
	}
	if isTag {
		value += fmt.Sprintf("tag:\n%s", getBlank())
	}
	if isMulti {
		value += fmt.Sprintf("multi:\n%s", getBlank())
	}
	if len(helpValues) > 0 {
		for _, v := range helpValues {
			value += fmt.Sprintf("val_help: %s\n", v)
		}
	}
	if allowed != nil {
		value += fmt.Sprintf("allowed: %s\n%s", *allowed, getBlank())
	}
	if syntax != nil {
		value += fmt.Sprintf("syntax:expression: %s\n%s", *syntax, getBlank())
	}

	return value
}

func createTestNode(
	isTag bool,
	isMulti bool,
	options []string,
) (*Node, error) {
	node := &Node{
		Name:        "TestNode",
		IsTag:       isTag,
		Multi:       isMulti,
		Children:    map[string]*Node{},
		Constraints: []NodeConstraint{},
	}

	var buffer bytes.Buffer
	for _, option := range options {
		buffer.Truncate(0)
		buffer.WriteString(option)

		err := parseDefinition(ioutil.NopCloser(&buffer), node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func createTestNormalNode(
	ntype *string,
	help *string,
	helpValues []string,
	allowed *string,
	syntax *string,
) (*Node, error) {
	return createTestNode(
		false,
		false,
		[]string{
			constructDefinition(
				ntype,
				help,
				false,
				false,
				helpValues,
				allowed,
				syntax,
			),
		})
}

func createTestTagNode(
	ntype *string,
	help *string,
	helpValues []string,
	allowed *string,
	syntax *string,
) (*Node, error) {
	return createTestNode(
		true,
		false,
		[]string{
			constructDefinition(
				ntype,
				help,
				true,
				false,
				helpValues,
				allowed,
				syntax,
			),
		})
}

func createTestMultiNode(
	ntype *string,
	help *string,
	helpValues []string,
	allowed *string,
	syntax *string,
) (*Node, error) {
	return createTestNode(
		false,
		true,
		[]string{
			constructDefinition(
				ntype,
				help,
				false,
				true,
				helpValues,
				allowed,
				syntax,
			),
		})
}

func createTestTagMultiNode(
	ntype *string,
	help *string,
	helpValues []string,
	allowed *string,
	syntax *string,
) (*Node, error) {
	return createTestNode(
		true,
		true,
		[]string{
			constructDefinition(
				ntype,
				help,
				true,
				true,
				helpValues,
				allowed,
				syntax,
			),
		})
}

func validateNode(node *Node, isTag bool, isMulti bool, ntype string, help string) []string {
	errors := []string{}

	if node.IsTag != isTag {
		errors = append(errors, fmt.Sprintf("Node tag should be %t, got %t", isTag, node.IsTag))
	}
	if node.Multi != isMulti {
		errors = append(errors, fmt.Sprintf("Node multi should be %t, got %t", isMulti, node.Multi))
	}
	if node.Type != ntype {
		errors = append(errors, fmt.Sprintf("Node should have type %s, got: '%s'", ntype, node.Type))
	}
	if node.Help != help {
		errors = append(errors, fmt.Sprintf("Node should have help %s, got: '%s'", help, node.Help))
	}

	return errors
}

func validateNormalNode(node *Node, ntype string, help string) []string {
	return validateNode(node, false, false, ntype, help)
}

func validateTagNode(node *Node, ntype string, help string) []string {
	return validateNode(node, true, false, ntype, help)
}

func validateMultiNode(node *Node, ntype string, help string) []string {
	return validateNode(node, false, true, ntype, help)
}

func validateTagMultiNode(node *Node, ntype string, help string) []string {
	return validateNode(node, true, true, ntype, help)
}

func validateConstraint(
	t *testing.T,
	node *Node,
	description string,
	constraintName ConstraintKey,
	value interface{},
	failureReason string,
) {
	if len(node.Constraints) != 1 {
		t.Errorf("Single constraint should be added for %s, got %d", description, len(node.Constraints))
		t.FailNow()
	}

	if !reflect.DeepEqual(node.Constraints[0].GetProperty(constraintName), value) {
		t.Errorf(
			"Constraint value for %s is incorrect, expected '%v' got '%v'",
			description,
			value,
			node.Constraints[0].GetProperty(constraintName),
		)
	}

	if node.Constraints[0].FailureReason != failureReason {
		t.Errorf(
			"Failure reason for constraint %s was incorrect, expected '%s' but got '%s'",
			description,
			failureReason,
			node.Constraints[0].FailureReason,
		)
	}
}

// TestSimpleDefinition checks basic type/help attributes for a node
func TestSimpleDefinition(t *testing.T) {
	ntype := "u32"
	help := "Port numbers to include in the group"
	val := "u32:1-65535 ;\\\nA port number to include"
	expr := ""

	node, err := createTestTagMultiNode(
		&ntype,
		&help,
		[]string{val},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagMultiNode(node, ntype, help) {
		t.Error(err)
	}

	if len(node.Constraints) > 0 {
		t.Error("No constraints expected")
	}

}

func TestConstraints(t *testing.T) {
	t.Run("Expression Bounds", testExprBounds)
	t.Run("Minimum Bound", testMinBound)
	t.Run("Maximum Bound", testMaxBound)
	t.Run("Pattern", testPattern)
	t.Run("Negated Pattern", testNegatedPattern)
	t.Run("Validate", testSimpleValidateCommand)
	t.Run("If-Block Validate", testIfBlockValidateCommand)
	t.Run("ExecNewline", testNewlineExec)
	t.Run("Expression List", testExprList)
	t.Run("Negated Expression List", testNegatedExprList)
	t.Run("Allowed CLI Shell", testAllowedCliShell)
	t.Run("Allowed Echo", testAllowedEcho)
	t.Run("Allowed Executable", testAllowedExecutable)
	t.Run("Allowed Array", testAllowedArrayVar)
	t.Run("Allowed Bash Command", testAllowedBashCommand)
	t.Run("Infinity/Number", testInfinityOrNumber)
	t.Run("Multi Option List", testMultiOptionList)
}

func testExprBounds(t *testing.T) {
	// See vpn/ipsec/esp-group/node.tag/proposal/node.def
	ntype := "txt"
	help := "Port number"
	reason := `Must be a valid port number`
	expr := fmt.Sprintf("($VAR(@) >= 1 && $VAR(@) <= 65535) ; \\\n    \"%s\"", reason)
	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Expression Bounds",
		MinBound,
		1,
		reason,
	)
	validateConstraint(
		t,
		node,
		"Expression Bounds",
		MaxBound,
		65535,
		reason,
	)
}

func testMinBound(t *testing.T) {
	// See vpn/ipsec/esp-group/node.tag/proposal/node.def
	ntype := "txt"
	help := "Port number"
	reason := `Must be a valid port number`
	expr := fmt.Sprintf("($VAR(@) > 0) ; \\\n    \"%s\"", reason)
	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Minimum bound",
		MinBound,
		1,
		reason,
	)
}

func testMaxBound(t *testing.T) {
	// See vpn/ipsec/esp-group/node.tag/proposal/node.def
	ntype := "txt"
	help := "Port number"
	reason := `Must be a valid port number`
	expr := fmt.Sprintf("$VAR(@) < 65536 ; \\\n    \"%s\"", reason)
	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Maximum bound",
		MaxBound,
		65535,
		reason,
	)
}

func testPattern(t *testing.T) {
	// See zone-policy/zone/node.def
	ntype := "txt"
	help := "Zone name"
	pattern := "^[[:print:]]{1,18}$"
	reason := `Zone name must be 18 characters or less`
	expr := fmt.Sprintf(`pattern $VAR(@) "%s" ;
                "%s"`, pattern, reason)
	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Pattern",
		Pattern,
		pattern,
		reason,
	)
}

func testNegatedPattern(t *testing.T) {
	// See service/webproxy/url-filtering/squidguard/local-ok/node.def
	// See zone-policy/zone/node.def
	ntype := "txt"
	help := "Local site to allow"
	pattern := "^https://"
	reason := `site should not start with https://`
	expr := fmt.Sprintf(`! pattern $VAR(@) "%s" ; \
                "%s"`, pattern, reason)
	node, err := createTestMultiNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateMultiNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Negated Pattern",
		NegatedPattern,
		pattern,
		reason,
	)
}

func testSimpleValidateCommand(t *testing.T) {
	// See firewall/name/node.def
	ntype := "txt"
	help := "Firewall name"
	command := `/usr/sbin/ubnt-fw validate-fw-name '$VAR(@)'`
	expr := fmt.Sprintf(`exec "%s"`, command)
	reason := ""

	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Simple Validate",
		ValidateCommand,
		command,
		reason,
	)
}

func testIfBlockValidateCommand(t *testing.T) {
	// See interfaces/name/node.tag/rule/node.tag/protocol/node.def
	ntype := "txt"
	help := "Interfaces on switch"
	command := `intf=$VAR(@); \
     port=${intf:3}; \
     if ! /usr/sbin/ubnt-hal onSwitch $port > /dev/null ; \
     then \
        echo \"interface $VAR(@): is not a switch port\"; \
        exit 1; \
     fi`
	expr := fmt.Sprintf(`exec \
	"%s"`, command)
	reason := ""

	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"If-Block Validate",
		ValidateCommand,
		command,
		reason,
	)
}

func testNewlineExec(t *testing.T) {
	// See system/ip/arp/table-size/node.def
	ntype := "u32"
	help := "Max number of entries to keep in the ARP cache"
	command := `/opt/vyatta/sbin/vyatta-update-arp-params       \
                'syntax-check' 'table-size' '$VAR(@)' 'ipv4'`
	expr := fmt.Sprintf(`exec "                               \
			%s"`, command)
	reason := ""

	node, err := createTestNormalNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNormalNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Validate with newline",
		ValidateCommand,
		command,
		reason,
	)
}

func testExprList(t *testing.T) {
	// See vpn/ipsec/logging/log-modes/node.def
	ntype := "txt"
	help := "Log mode, reference strongSwan documentation"
	options := []string{
		"dmn", "mgr", "ike", "chd", "job", "cfg", "knl", "net", "asn", "enc", "lib", "esp", "tls", "tnc", "imc", "imv", "pts",
	}
	reason := `must be one of the following: dmn, mgr, ike, chd, job, cfg, knl, net, asn, enc, lib, esp, tls, tnc, imc, imv, pts`
	expr := fmt.Sprintf(
		`$VAR(@) in "%s"; "%s"`,
		strings.Join(options, `", "`),
		reason,
	)

	node, err := createTestMultiNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateMultiNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Expression List",
		Options,
		options,
		reason,
	)
}

func testNegatedExprList(t *testing.T) {
	// See system/login/user/node.tag/group/node.def
	ntype := "txt"
	help := "Additional group membership"
	options := []string{
		"quaggavty", "vyattacfg", "vyattaop", "sudo", "adm", "operator",
	}
	reason := "Use configuration level to change membership of operator and admin groups"
	expr := fmt.Sprintf(
		`! $VAR(@) in \
			"%s"
			; "%s"`,
		strings.Join(options, `", "`),
		reason,
	)

	node, err := createTestMultiNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateMultiNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Negated Expression List",
		NegatedOptions,
		options,
		reason,
	)
}

func testAllowedCliShell(t *testing.T) {
	// See vpn/ipsec/remote-access/ike-settings/esp-group/node.def
	ntype := "txt"
	help := "Default ESP group name"
	command := `cli-shell-api listActiveNodes vpn ipsec esp-group`
	expr := fmt.Sprintf(`allowed: %s`, command)
	reason := ""

	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Allowed CLI Shell",
		OptionsCommand,
		command,
		reason,
	)
}

func testAllowedEcho(t *testing.T) {
	// See system/ip/arp/table-size/node.deF
	ntype := "u32"
	help := "Max number of entries to keep in the ARP cache"
	command := `echo "1024 2048 4096 8192 16384 32768 65536 131072 262144"`
	expr := fmt.Sprintf(`allowed: %s`, command)
	reason := ""

	node, err := createTestNormalNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNormalNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Allowed Echo",
		OptionsCommand,
		command,
		reason,
	)
}

func testAllowedExecutable(t *testing.T) {
	// See interfaces/switch/node.tag/redirect/node.def
	ntype := "txt"
	help := "Incoming packet redirection destination"
	command := `/opt/vyatta/sbin/ubnt-interface --show=input`
	expr := fmt.Sprintf(`allowed: %s`, command)
	reason := ""

	node, err := createTestNormalNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNormalNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Allowed Executable",
		OptionsCommand,
		command,
		reason,
	)
}

func testAllowedArrayVar(t *testing.T) {
	// See interfaces/switch/node.tag/switch-port/interface/node.def
	ntype := "txt"
	help := "Interfaces on switch"
	command := `local -a array
         ports=` + "`/usr/sbin/ubnt-hal getPortCount`" + `
         (( ports-- ))
         for ((i=0; i<=$ports; i++)); do
            if /usr/sbin/ubnt-hal onSwitch $i > /dev/null ; then
               array+=(eth$i)
            fi
         done
         echo -n ${array[@]##*/}`
	expr := fmt.Sprintf(`allowed: %s`, command)
	reason := ""

	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Allowed Array",
		OptionsCommand,
		command,
		reason,
	)
}

func testAllowedBashCommand(t *testing.T) {
	// See firewall/name/node.tag/rule/node.tag/protocol/node.def
	ntype := "txt"
	help := "Protocol to match"
	command := `protos=` + "cat /etc/protocols | sed -e '/^#.*/d' | awk '{ print $1 }' | grep -v 'v6'`" + `
        protos="all $protos tcp_udp"
        echo -n $protos`
	expr := fmt.Sprintf(`allowed:
	%s`, command)
	reason := ""

	node, err := createTestTagNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateTagNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Allowed Bash Command",
		OptionsCommand,
		command,
		reason,
	)
}

func testInfinityOrNumber(t *testing.T) {
	// See interfaces/bridge/node.tag/ipv6/router-advert/prefix/node.tag/valid-lifetime/node.def
	// See firewall/name/node.tag/rule/node.tag/protocol/node.def
	ntype := "txt"
	help := "Time in seconds the prefix will remain valid"
	option := "infinity"
	pattern := "[0-9]*"
	reason := "Must be 'infinity' or a number"
	expr := fmt.Sprintf(
		`($VAR(@) == "%s" || (pattern $VAR(@) "%s")) ; "%s"`,
		option,
		pattern,
		reason,
	)

	node, err := createTestNormalNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNormalNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Infinity/Number",
		Options,
		[]string{option},
		reason,
	)
	validateConstraint(
		t,
		node,
		"Infinity/Number",
		Pattern,
		pattern,
		reason,
	)

}

func testMultiOptionList(t *testing.T) {
	// See service/dhcp-server/use-dnsmasq/node.def
	ntype := "txt"
	help := "Option to use dnsmasq as DHCP server"
	options := []string{
		"disable", "enable",
	}
	reason := `"use-dnsmasq" must be "enable" or "disable"`
	expr := `($VAR(@) == "disable" || $VAR(@) == "enable" ); \
    "\"use-dnsmasq\" must be \"enable\" or \"disable\""`

	node, err := createTestNormalNode(
		&ntype,
		&help,
		[]string{},
		nil,
		&expr,
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNormalNode(node, ntype, help) {
		t.Error(err)
	}

	validateConstraint(
		t,
		node,
		"Multi Option List",
		Options,
		options,
		reason,
	)
}
