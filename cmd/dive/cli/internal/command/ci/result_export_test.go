package ci

// StatusString exposes the rule status (without styling) to the external golden test.
func (r RuleResult) StatusString() string { return r.status.String(format{}) }

// Message exposes the rule message to the external golden test.
func (r RuleResult) Message() string { return r.message }
