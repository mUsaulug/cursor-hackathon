package vision

// normalize maps a raw model label to a normalized urban object type using the
// ontology rules. The bool reports whether the label was explicitly mapped;
// false means it fell back to the ontology's fallback type (unknown) and will
// be routed to human review regardless of confidence (design doc 5.1, 5.6).
func (p *Pipeline) normalize(label string) (string, bool) {
	return p.rules.Ontology.Normalize(label)
}
