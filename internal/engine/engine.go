package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Cunyule/AI-coding/internal/model"
)

type OSRule struct {
	Pattern string `json:"pattern"`
	Value   string `json:"value"`
}

type Rule struct {
	ID             string   `json:"id"`
	Priority       int      `json:"priority"`
	Protocol       string   `json:"protocol"`
	Product        string   `json:"product"`
	Pattern        string   `json:"pattern"`
	VersionGroup   string   `json:"version_group"`
	Ports          []int    `json:"ports"`
	BaseConfidence float64  `json:"base_confidence"`
	OSRules        []OSRule `json:"os_rules"`
}

type compiledOSRule struct {
	pattern *regexp.Regexp
	value   string
}

type compiledRule struct {
	rule    Rule
	pattern *regexp.Regexp
	osRules []compiledOSRule
}

type Engine struct {
	rules []compiledRule
}

func Load(path string) (*Engine, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules: %w", err)
	}
	var rules []Rule
	if err := json.Unmarshal(b, &rules); err != nil {
		return nil, fmt.Errorf("decode rules: %w", err)
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("no fingerprint rules configured")
	}
	compiled := make([]compiledRule, 0, len(rules))
	for _, rule := range rules {
		if rule.ID == "" || rule.Protocol == "" || rule.Pattern == "" {
			return nil, fmt.Errorf("rule requires id, protocol and pattern")
		}
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return nil, fmt.Errorf("compile rule %q: %w", rule.ID, err)
		}
		cr := compiledRule{rule: rule, pattern: re}
		for _, osRule := range rule.OSRules {
			osRE, err := regexp.Compile(osRule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("compile os rule in %q: %w", rule.ID, err)
			}
			cr.osRules = append(cr.osRules, compiledOSRule{pattern: osRE, value: osRule.Value})
		}
		compiled = append(compiled, cr)
	}
	sort.SliceStable(compiled, func(i, j int) bool { return compiled[i].rule.Priority > compiled[j].rule.Priority })
	return &Engine{rules: compiled}, nil
}

func (e *Engine) Identify(input model.ScanRecord) model.FingerprintResult {
	result := model.FingerprintResult{IP: input.IP, Port: input.Port, Protocol: "unknown"}
	bestScore := -1.0
	bestPriority := -1

	for _, candidate := range e.rules {
		match := candidate.pattern.FindStringSubmatch(input.Banner)
		if match == nil {
			continue
		}
		score := candidate.rule.BaseConfidence
		if containsPort(candidate.rule.Ports, input.Port) {
			score += 0.03
		}
		if score < bestScore || (score == bestScore && candidate.rule.Priority <= bestPriority) {
			continue
		}
		version := ""
		if candidate.rule.VersionGroup != "" {
			index := candidate.pattern.SubexpIndex(candidate.rule.VersionGroup)
			if index > 0 && index < len(match) {
				version = strings.TrimSpace(match[index])
			}
		}
		osHint := ""
		for _, osRule := range candidate.osRules {
			if osRule.pattern.MatchString(input.Banner) {
				osHint = osRule.value
				break
			}
		}
		result.Protocol = candidate.rule.Protocol
		result.Product = candidate.rule.Product
		result.Version = version
		result.OSHint = osHint
		result.Confidence = math.Round(math.Min(1, score)*100) / 100
		bestScore = score
		bestPriority = candidate.rule.Priority
	}
	return result
}

func (e *Engine) IdentifyBatch(inputs []model.ScanRecord) []model.FingerprintResult {
	results := make([]model.FingerprintResult, len(inputs))
	for i, input := range inputs {
		results[i] = e.Identify(input)
	}
	return results
}

func containsPort(ports []int, port int) bool {
	for _, candidate := range ports {
		if candidate == port {
			return true
		}
	}
	return false
}
