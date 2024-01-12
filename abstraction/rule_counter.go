package abstraction

import "sync"

// Needs to be an invalid firewall name to ensure no collisions
var NAT_COUNTER = "#service_nat_rules#"

// RuleCounter keeps track of the current rule number to use for a certain firewall/NAT
type RuleCounter struct {
	number int
	step   int
	lock   *sync.Mutex
}

func (counter RuleCounter) Next() int {
	counter.lock.Lock()
	defer counter.lock.Unlock()
	defer func() { counter.number += counter.step }()

	return counter.number
}

var counters map[string]RuleCounter

func MakeCounter(name string, ruleStart int, ruleStep int) RuleCounter {
	counters[name] = RuleCounter{
		number: ruleStart,
		step:   ruleStep,
		lock:   &sync.Mutex{},
	}

	return counters[name]
}

func HasCounter(name string) bool {
	_, ok := counters[name]
	return ok
}

// GetCounter will return a rule counter, or create one if not existing
func GetCounter(name string) RuleCounter {
	return counters[name]
}
