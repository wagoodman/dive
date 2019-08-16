package runtime

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/logrusorgru/aurora"
	"github.com/wagoodman/dive/image"
)

type CiEvaluator struct {
	Rules            []CiRule
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

func NewCiEvaluator(config *viper.Viper) *CiEvaluator {
	return &CiEvaluator{
		Rules:   loadCiRules(config),
		Results: make(map[string]RuleResult),
		Pass:    true,
	}
}

func (ci *CiEvaluator) isRuleEnabled(rule CiRule) bool {
	return rule.Configuration() != "disabled"
}

func (ci *CiEvaluator) Evaluate(analysis *image.AnalysisResult) bool {
	canEvaluate := true
	for _, rule := range ci.Rules {
		if !ci.isRuleEnabled(rule) {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleConfigured,
				message: "rule disabled",
			}
			continue
		}

		err := rule.Validate()
		if err != nil {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleMisconfigured,
				message: err.Error(),
			}
			canEvaluate = false
		} else {
			ci.Results[rule.Key()] = RuleResult{
				status:  RuleConfigured,
				message: "test",
			}
		}

	}

	if !canEvaluate {
		ci.Pass = false
		ci.Misconfigured = true
		return ci.Pass
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

func (ci *CiEvaluator) Report() {
	fmt.Println(title("Inefficient Files:"))

	template := "%5s  %12s  %-s\n"
	fmt.Printf(template, "Count", "Wasted Space", "File Path")

	if len(ci.InefficientFiles) == 0 {
		fmt.Println("None")
	} else {
		for _, file := range ci.InefficientFiles {
			fmt.Printf(template, strconv.Itoa(file.References), humanize.Bytes(uint64(file.SizeBytes)), file.Path)
		}
	}

	fmt.Println(title("Results:"))

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
		name := strings.TrimPrefix(rule, "rules.")
		if result.message != "" {
			fmt.Printf("  %s: %s: %s\n", result.status.String(), name, result.message)
		} else {
			fmt.Printf("  %s: %s\n", result.status.String(), name)
		}
	}

	if ci.Misconfigured {
		fmt.Println(aurora.Red("CI Misconfigured"))
		return
	}

	summary := fmt.Sprintf("Result:%s [Total:%d] [Passed:%d] [Failed:%d] [Warn:%d] [Skipped:%d]", status, ci.Tally.Total, ci.Tally.Pass, ci.Tally.Fail, ci.Tally.Warn, ci.Tally.Skip)
	if ci.Pass {
		fmt.Println(aurora.Green(summary))
	} else if ci.Pass && ci.Tally.Warn > 0 {
		fmt.Println(aurora.Blue(summary))
	} else {
		fmt.Println(aurora.Red(summary))
	}
}
