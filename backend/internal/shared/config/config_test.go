package config

import "testing"

func TestParseYAMLSubset(t *testing.T) {
	src := []byte(`
version: "1.0"
# a comment line
amap:
  traffic light: traffic_signal
  D40: road_damage
alist:
  - person
  - bicycle
scalar_num: 0.85
`)
	doc, err := parseYAML(src)
	if err != nil {
		t.Fatalf("parseYAML: %v", err)
	}
	if v, _ := doc.scalar("version"); v != "1.0" {
		t.Errorf("version = %q, want 1.0", v)
	}
	if doc.maps["amap"]["traffic light"] != "traffic_signal" {
		t.Errorf("amap[traffic light] = %q", doc.maps["amap"]["traffic light"])
	}
	if doc.maps["amap"]["D40"] != "road_damage" {
		t.Errorf("amap[D40] = %q", doc.maps["amap"]["D40"])
	}
	if got := doc.lists["alist"]; len(got) != 2 || got[0] != "person" || got[1] != "bicycle" {
		t.Errorf("alist = %v", got)
	}
	if f, ok, err := doc.float("scalar_num"); err != nil || !ok || f != 0.85 {
		t.Errorf("scalar_num = %v ok=%v err=%v", f, ok, err)
	}
}

func TestParseYAMLRejectsMixedBlock(t *testing.T) {
	src := []byte("amap:\n  key: value\n  - item\n")
	if _, err := parseYAML(src); err == nil {
		t.Fatal("expected error for list item inside map block")
	}
}

func TestLoadRulesFromEmbed(t *testing.T) {
	r, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Ontology: real labels map to expected types.
	if got, ok := r.Ontology.Normalize("traffic light"); !ok || got != "traffic_signal" {
		t.Errorf("Normalize(traffic light) = %q ok=%v", got, ok)
	}
	if got, ok := r.Ontology.Normalize("D40"); !ok || got != "road_damage" {
		t.Errorf("Normalize(D40) = %q ok=%v", got, ok)
	}
	if got, ok := r.Ontology.Normalize("banana"); ok || got != "unknown" {
		t.Errorf("Normalize(banana) = %q ok=%v, want unknown,false", got, ok)
	}

	// KVKK policy.
	if !r.Ontology.IsBlocked("person") {
		t.Error("person should be blocked")
	}
	if !r.Ontology.IsHidden("car") {
		t.Error("car should be hidden")
	}
	if r.Ontology.IsBlocked("traffic light") {
		t.Error("traffic light must not be blocked")
	}

	// Priority policy.
	if got := r.Priority.Priority("road_damage"); got != "high" {
		t.Errorf("priority road_damage = %q, want high", got)
	}
	if got := r.Priority.Priority("street_furniture"); got != "low" {
		t.Errorf("priority street_furniture = %q, want low", got)
	}
	if got := r.Priority.Priority("nonexistent"); got != "low" {
		t.Errorf("priority default = %q, want low", got)
	}

	// Confidence thresholds.
	if r.Confidence.DefaultAutoAccept != 0.80 {
		t.Errorf("default auto_accept = %v, want 0.80", r.Confidence.DefaultAutoAccept)
	}
	if r.Confidence.DefaultNeedsReview != 0.50 {
		t.Errorf("default needs_review = %v, want 0.50", r.Confidence.DefaultNeedsReview)
	}
	if got := r.Confidence.AutoAccept("road_damage"); got != 0.85 {
		t.Errorf("auto_accept road_damage = %v, want 0.85", got)
	}
	if got := r.Confidence.AutoAccept("nonexistent"); got != 0.80 {
		t.Errorf("auto_accept fallback = %v, want 0.80", got)
	}
}
