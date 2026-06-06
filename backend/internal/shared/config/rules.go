package config

import (
	"embed"
	"fmt"
)

//go:embed rules/ontology.yaml rules/priority_policy.yaml rules/confidence_thresholds.yaml
var rulesFS embed.FS

// Ontology maps raw model labels to normalized urban object types and holds the
// KVKK label policy. Source: rules/ontology.yaml.
type Ontology struct {
	Version       string
	COCO          map[string]string // COCO label -> normalized type
	RDD           map[string]string // RDD class -> normalized type
	FallbackType  string
	BlockedLabels map[string]bool // removed entirely (KVKK)
	HideLabels    map[string]bool // excluded by default (tracking/plate risk)
}

// Normalize returns the normalized object type for a raw label and whether the
// label was explicitly mapped (false => fell back to FallbackType).
func (o Ontology) Normalize(label string) (string, bool) {
	if t, ok := o.COCO[label]; ok {
		return t, true
	}
	if t, ok := o.RDD[label]; ok {
		return t, true
	}
	return o.FallbackType, false
}

// IsBlocked reports whether a raw label must be dropped for KVKK reasons.
func (o Ontology) IsBlocked(label string) bool { return o.BlockedLabels[label] }

// IsHidden reports whether a raw label is hidden by default policy.
func (o Ontology) IsHidden(label string) bool { return o.HideLabels[label] }

// PriorityPolicy maps normalized type -> priority string. Source:
// rules/priority_policy.yaml.
type PriorityPolicy struct {
	Version         string
	PriorityByType  map[string]string
	DefaultPriority string
}

// Priority returns the configured priority for a normalized type.
func (p PriorityPolicy) Priority(objectType string) string {
	if v, ok := p.PriorityByType[objectType]; ok {
		return v
	}
	return p.DefaultPriority
}

// ConfidenceThresholds drives the confidence evaluator and review router.
// Source: rules/confidence_thresholds.yaml.
type ConfidenceThresholds struct {
	Version            string
	DefaultAutoAccept  float64
	DefaultNeedsReview float64
	AutoAcceptByType   map[string]float64
}

// AutoAccept returns the auto-accept threshold for a normalized type, falling
// back to the default when no per-type override exists.
func (c ConfidenceThresholds) AutoAccept(objectType string) float64 {
	if v, ok := c.AutoAcceptByType[objectType]; ok {
		return v
	}
	return c.DefaultAutoAccept
}

// Rules bundles all loaded rule sets.
type Rules struct {
	Ontology   Ontology
	Priority   PriorityPolicy
	Confidence ConfidenceThresholds
}

// Load reads and parses all embedded rule files. It fails fast: a malformed or
// missing rule file is a programming error, not a runtime condition.
func Load() (*Rules, error) {
	ont, err := loadOntology()
	if err != nil {
		return nil, err
	}
	pri, err := loadPriority()
	if err != nil {
		return nil, err
	}
	conf, err := loadConfidence()
	if err != nil {
		return nil, err
	}
	return &Rules{Ontology: ont, Priority: pri, Confidence: conf}, nil
}

// MustLoad is Load but panics on error. Use at startup wiring.
func MustLoad() *Rules {
	r, err := Load()
	if err != nil {
		panic(err)
	}
	return r
}

func readRule(name string) (*yamlDoc, error) {
	data, err := rulesFS.ReadFile("rules/" + name)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", name, err)
	}
	doc, err := parseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", name, err)
	}
	return doc, nil
}

func loadOntology() (Ontology, error) {
	doc, err := readRule("ontology.yaml")
	if err != nil {
		return Ontology{}, err
	}
	o := Ontology{
		COCO:          doc.maps["coco"],
		RDD:           doc.maps["rdd"],
		BlockedLabels: toSet(doc.lists["blocked_labels"]),
		HideLabels:    toSet(doc.lists["hide_labels"]),
	}
	o.Version, _ = doc.scalar("version")
	o.FallbackType, _ = doc.scalar("fallback_type")
	if o.FallbackType == "" {
		o.FallbackType = "unknown"
	}
	if o.COCO == nil {
		o.COCO = map[string]string{}
	}
	if o.RDD == nil {
		o.RDD = map[string]string{}
	}
	return o, nil
}

func loadPriority() (PriorityPolicy, error) {
	doc, err := readRule("priority_policy.yaml")
	if err != nil {
		return PriorityPolicy{}, err
	}
	p := PriorityPolicy{PriorityByType: doc.maps["priority_by_type"]}
	p.Version, _ = doc.scalar("version")
	p.DefaultPriority, _ = doc.scalar("default_priority")
	if p.DefaultPriority == "" {
		p.DefaultPriority = "low"
	}
	if p.PriorityByType == nil {
		p.PriorityByType = map[string]string{}
	}
	return p, nil
}

func loadConfidence() (ConfidenceThresholds, error) {
	doc, err := readRule("confidence_thresholds.yaml")
	if err != nil {
		return ConfidenceThresholds{}, err
	}
	c := ConfidenceThresholds{AutoAcceptByType: map[string]float64{}}
	c.Version, _ = doc.scalar("version")

	if v, ok, err := doc.float("default_auto_accept"); err != nil {
		return ConfidenceThresholds{}, err
	} else if ok {
		c.DefaultAutoAccept = v
	} else {
		c.DefaultAutoAccept = 0.80
	}
	if v, ok, err := doc.float("default_needs_review"); err != nil {
		return ConfidenceThresholds{}, err
	} else if ok {
		c.DefaultNeedsReview = v
	} else {
		c.DefaultNeedsReview = 0.50
	}
	for k, raw := range doc.maps["auto_accept_by_type"] {
		f, err := parseFloat(raw)
		if err != nil {
			return ConfidenceThresholds{}, fmt.Errorf("config: auto_accept_by_type[%s]: %w", k, err)
		}
		c.AutoAcceptByType[k] = f
	}
	return c, nil
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		if it != "" {
			m[it] = true
		}
	}
	return m
}
