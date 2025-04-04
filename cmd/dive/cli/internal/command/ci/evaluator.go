package ci

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/context"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wagoodman/dive/dive/image"
)

type Evaluation struct {
	Report string
	Pass   bool
}

type Evaluator struct {
	Rules            []Rule
	Results          map[string]RuleResult
	Tally            ResultTally
	Pass             bool
	Misconfigured    bool
	InefficientFiles []ReferenceFile
	format           format
}

type format struct {
	Title       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Disabled    lipgloss.Style
	Failure     lipgloss.Style
	TableHeader lipgloss.Style
	Label       lipgloss.Style
	Aux         lipgloss.Style
	Value       lipgloss.Style
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

func NewEvaluator(rules []Rule) Evaluator {
	return Evaluator{
		Rules:   rules,
		Results: make(map[string]RuleResult),
		Pass:    true,
		format: format{
			Title:       lipgloss.NewStyle().Bold(true),
			Success:     lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
			Warning:     lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
			Disabled:    lipgloss.NewStyle().Faint(true),
			Failure:     lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Reverse(true),
			TableHeader: lipgloss.NewStyle().Bold(true).Underline(true),
			Label:       lipgloss.NewStyle().Width(18),
			Aux:         lipgloss.NewStyle().Faint(true),
			Value:       lipgloss.NewStyle(),
		},
	}
}

func (e Evaluator) isRuleEnabled(rule Rule) bool {
	return rule.Configuration() != "disabled"
}

func (e Evaluator) Evaluate(ctx context.Context, analysis *image.Analysis) Evaluation {
	for _, rule := range e.Rules {
		if !e.isRuleEnabled(rule) {
			e.Results[rule.Key()] = RuleResult{
				status:  RuleConfigured,
				message: "rule disabled",
			}
			continue
		}

		e.Results[rule.Key()] = RuleResult{
			status:  RuleConfigured,
			message: "test",
		}
	}

	// capture inefficient files
	for idx := 0; idx < len(analysis.Inefficiencies); idx++ {
		fileData := analysis.Inefficiencies[len(analysis.Inefficiencies)-1-idx]

		e.InefficientFiles = append(e.InefficientFiles, ReferenceFile{
			References: len(fileData.Nodes),
			SizeBytes:  uint64(fileData.CumulativeSize),
			Path:       fileData.Path,
		})
	}

	// evaluate results against the configured CI rules
	for _, rule := range e.Rules {
		if !e.isRuleEnabled(rule) {
			e.Results[rule.Key()] = RuleResult{
				status:  RuleDisabled,
				message: "disabled",
			}
			continue
		}

		status, message := rule.Evaluate(analysis)

		if value, exists := e.Results[rule.Key()]; exists && value.status != RuleConfigured && value.status != RuleMisconfigured {
			panic(fmt.Errorf("CI rule result recorded twice: %s", rule.Key()))
		}

		if status == RuleFailed {
			e.Pass = false
		}

		if message == "" {
			message = rule.Configuration()
		}

		e.Results[rule.Key()] = RuleResult{
			status:  status,
			message: message,
		}
	}

	e.Tally.Total = len(e.Results)
	for rule, result := range e.Results {
		switch result.status {
		case RulePassed:
			e.Tally.Pass++
		case RuleFailed:
			e.Tally.Fail++
		case RuleWarning:
			e.Tally.Warn++
		case RuleDisabled:
			e.Tally.Skip++
		default:
			panic(fmt.Errorf("unknown test status (rule='%v'): %v", rule, result.status))
		}
	}

	return Evaluation{
		Report: e.report(analysis),
		Pass:   e.Pass,
	}
}

func (e Evaluator) report(analysis *image.Analysis) string {
	sections := []string{
		e.renderAnalysisSection(analysis),
		e.renderInefficientFilesSection(analysis),
		e.renderEvaluationSection(),
	}

	return strings.Join(sections, "\n\n")
}

func (e Evaluator) renderAnalysisSection(analysis *image.Analysis) string {
	wastedByteStr := ""
	userWastedPercent := "0 %"

	if analysis.WastedBytes > 0 {
		wastedByteStr = fmt.Sprintf("(%s)", humanize.Bytes(analysis.WastedBytes))
		userWastedPercent = fmt.Sprintf("%.2f %%", analysis.WastedUserPercent*100)
	}

	title := e.format.Title.Render("Analysis:")

	rows := []string{
		formatKeyValue(e.format, "efficiency", fmt.Sprintf("%.2f %%", analysis.Efficiency*100)),
		formatKeyValue(e.format, "wastedBytes", fmt.Sprintf("%d bytes %s", analysis.WastedBytes, wastedByteStr)),
		formatKeyValue(e.format, "userWastedPercent", userWastedPercent),
	}

	return title + "\n" + strings.Join(rows, "\n")
}

func (e Evaluator) renderInefficientFilesSection(analysis *image.Analysis) string {
	title := e.format.Title.Render("Inefficient Files:")

	if len(analysis.Inefficiencies) == 0 {
		return title + " (None)"
	}

	header := e.format.TableHeader.Render(
		fmt.Sprintf("%-5s  %-12s  %-s", "Count", "Wasted Space", "File Path"),
	)

	rows := []string{header}
	for _, file := range e.InefficientFiles {
		row := fmt.Sprintf("%-5s  %-12s  %-s",
			strconv.Itoa(file.References),
			humanize.Bytes(file.SizeBytes),
			file.Path,
		)
		rows = append(rows, row)
	}

	return title + "\n" + strings.Join(rows, "\n")
}

func (e Evaluator) renderEvaluationSection() string {
	title := e.format.Title.Render("Evaluation:")

	// sort rules by name for consistent output
	rules := make([]string, 0, len(e.Results))
	for name := range e.Results {
		rules = append(rules, name)
	}
	sort.Strings(rules)

	ruleResults := []string{}
	for _, rule := range rules {
		result := e.Results[rule]
		ruleResult := e.formatRuleResult(rule, result)
		ruleResults = append(ruleResults, ruleResult)
	}

	status := e.renderStatusSummary()

	return title + "\n" + strings.Join(ruleResults, "\n") + "\n\n" + status
}

func (e Evaluator) formatRuleResult(ruleName string, result RuleResult) string {
	var style lipgloss.Style
	textStyle := lipgloss.NewStyle()
	switch result.status {
	case RulePassed:
		style = e.format.Success
	case RuleFailed:
		style = e.format.Failure
	case RuleWarning, RuleMisconfigured:
		style = e.format.Warning
	case RuleDisabled:
		style = e.format.Disabled
		textStyle = e.format.Disabled
	default:
		style = lipgloss.NewStyle()
	}

	statusStr := style.Render(result.status.String(e.format))

	if result.message != "" {
		return fmt.Sprintf("  %s  %s", statusStr, textStyle.Render(ruleName+" ("+result.message+")"))
	}

	return fmt.Sprintf("  %s  %s", statusStr, textStyle.Render(ruleName))
}

func (e Evaluator) renderStatusSummary() string {
	if e.Misconfigured {
		return e.format.Failure.Render("CI Misconfigured")
	}

	status := "PASS"
	if e.Tally.Fail > 0 {
		status = "FAIL"
	}

	parts := []string{}

	type tallyItem struct {
		name  string
		value int
	}

	items := []tallyItem{
		//{"total", e.Tally.Total},
		{"pass", e.Tally.Pass},
		{"fail", e.Tally.Fail},
		{"warn", e.Tally.Warn},
		{"skip", e.Tally.Skip},
	}

	for _, item := range items {
		if item.value > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", item.name, item.value))
		}
	}

	auxSummary := e.format.Aux.Render(" [" + strings.Join(parts, " ") + "]")

	var style lipgloss.Style
	switch {
	case e.Pass && e.Tally.Warn == 0:
		style = e.format.Success
	case e.Pass && e.Tally.Warn > 0:
		style = e.format.Warning
	default:
		style = e.format.Failure
	}
	return style.Render(status) + auxSummary
}

func formatKeyValue(f format, key, value string) string {
	formattedKey := f.Label.Render(key + ":")
	return fmt.Sprintf("  %s %s", formattedKey, value)
}
