package ci

import (
	"fmt"
	"github.com/wagoodman/dive/internal/utils"
	"golang.org/x/net/context"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/logrusorgru/aurora/v4"
	"github.com/wagoodman/dive/dive/image"
)

type Evaluator struct {
	Rules            []Rule
	Results          map[string]RuleResult
	Tally            ResultTally
	Pass             bool
	Misconfigured    bool
	InefficientFiles []ReferenceFile
}

type ResultTally struct {
	Pass  int
	Fail  int
	Skip  int
	Warn  int
	Total int
}

type ReferenceFile struct {
	References int    `json:"count"`
	SizeBytes  uint64 `json:"sizeBytes"`
	Path       string `json:"file"`
}

func NewEvaluator(rules []Rule) *Evaluator {
	return &Evaluator{
		Rules:   rules,
		Results: make(map[string]RuleResult),
		Pass:    true,
	}
}

func (ci *Evaluator) isRuleEnabled(rule Rule) bool {
	return rule.Configuration() != "disabled"
}

func (ci *Evaluator) Evaluate(ctx context.Context, analysis *image.Analysis) bool {
	for _, rule := range ci.Rules {
		if !ci.isRuleEnabled(rule) {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleConfigured,
				message: "rule disabled",
			}
			continue
		}

		ci.Results[rule.Key()] = RuleResult{
			status:  RuleConfigured,
			message: "test",
		}
	}

	// capture inefficient files
	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]

		ci.InefficientFiles = append(ci.InefficientFiles, ReferenceFile{
			References: len(fileData.Nodes),
			SizeBytes:  uint64(fileData.CumulativeSize),
			Path:       fileData.Path,
		})
	}

	// evaluate results against the configured CI rules
	for _, rule := range ci.Rules {
		if !ci.isRuleEnabled(rule) {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleDisabled,
				message: "rule disabled",
			}
			continue
		}

		status, message := rule.Evaluate(analysis)

		if value, exists := ci.Results[rule.Key()]; exists && value.status != RuleConfigured && value.status != RuleMisconfigured {
			panic(fmt.Errorf("CI rule result recorded twice: %s", rule.Key()))
		}

		if status == RuleFailed {
			ci.Pass = false
		}

		ci.Results[rule.Key()] = RuleResult{
			status:  status,
			message: message,
		}
	}

	ci.Tally.Total = len(ci.Results)
	for rule, result := range ci.Results {
		switch result.status {
		case RulePassed:
			ci.Tally.Pass++
		case RuleFailed:
			ci.Tally.Fail++
		case RuleWarning:
			ci.Tally.Warn++
		case RuleDisabled:
			ci.Tally.Skip++
		default:
			panic(fmt.Errorf("unknown test status (rule='%v'): %v", rule, result.status))
		}
	}

	return ci.Pass
}

func (ci *Evaluator) Report() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, utils.TitleFormat("Inefficient Files:"))

	template := "%5s  %12s  %-s\n"
	fmt.Fprintf(&sb, template, "Count", "Wasted Space", "File Path")

	if len(ci.InefficientFiles) == 0 {
		fmt.Fprintln(&sb, "None")
	} else {
		for _, file := range ci.InefficientFiles {
			fmt.Fprintf(&sb, template, strconv.Itoa(file.References), humanize.Bytes(file.SizeBytes), file.Path)
		}
	}

	fmt.Fprintln(&sb, utils.TitleFormat("Results:"))

	status := "PASS"

	rules := make([]string, 0, len(ci.Results))
	for name := range ci.Results {
		rules = append(rules, name)
	}
	sort.Strings(rules)

	if ci.Tally.Fail > 0 {
		status = "FAIL"
	}

	for _, rule := range rules {
		result := ci.Results[rule]
		if result.message != "" {
			fmt.Fprintf(&sb, "  %s: %s: %s\n", result.status.String(), rule, result.message)
		} else {
			fmt.Fprintf(&sb, "  %s: %s\n", result.status.String(), rule)
		}
	}

	if ci.Misconfigured {
		fmt.Fprintln(&sb, aurora.Red("CI Misconfigured"))
	} else {
		summary := fmt.Sprintf("Result:%s [Total:%d] [Passed:%d] [Failed:%d] [Warn:%d] [Skipped:%d]", status, ci.Tally.Total, ci.Tally.Pass, ci.Tally.Fail, ci.Tally.Warn, ci.Tally.Skip)
		if ci.Pass {
			fmt.Fprintln(&sb, aurora.Green(summary))
		} else if ci.Pass && ci.Tally.Warn > 0 {
			fmt.Fprintln(&sb, aurora.Blue(summary))
		} else {
			fmt.Fprintln(&sb, aurora.Red(summary))
		}
	}
	return sb.String()
}
