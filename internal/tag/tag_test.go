package tag

import (
	"testing"
)

func TestNew_ParsesValidEntries(t *testing.T) {
	tgr := New([]string{"env=prod", "region=us-east-1", "team=platform"})
	tags := tgr.Tags()
	if tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", tags["env"])
	}
	if tags["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", tags["region"])
	}
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
}

func TestNew_IgnoresInvalidEntries(t *testing.T) {
	tgr := New([]string{"no-equals", "=emptykey", "valid=yes"})
	tags := tgr.Tags()
	if _, ok := tags[""]; ok {
		t.Error("empty key should be ignored")
	}
	if tags["valid"] != "yes" {
		t.Errorf("expected valid=yes, got %q", tags["valid"])
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestNew_EmptySlice_IsEmpty(t *testing.T) {
	tgr := New(nil)
	if !tgr.Empty() {
		t.Error("expected tagger to be empty")
	}
}

func TestApply_MergesIntoDestination(t *testing.T) {
	tgr := New([]string{"env=prod", "region=us-east-1"})
	dst := map[string]string{"job": "backup"}
	result := tgr.Apply(dst)
	if result["job"] != "backup" {
		t.Error("existing key should be preserved")
	}
	if result["env"] != "prod" {
		t.Error("tag env should be added")
	}
	if result["region"] != "us-east-1" {
		t.Error("tag region should be added")
	}
}

func TestApply_DoesNotOverwriteExistingKeys(t *testing.T) {
	tgr := New([]string{"env=prod"})
	dst := map[string]string{"env": "staging"}
	result := tgr.Apply(dst)
	if result["env"] != "staging" {
		t.Errorf("existing key should not be overwritten, got %q", result["env"])
	}
}

func TestApply_NilDestination_ReturnsNewMap(t *testing.T) {
	tgr := New([]string{"env=prod"})
	result := tgr.Apply(nil)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if result["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", result["env"])
	}
}

func TestTags_ReturnsCopy(t *testing.T) {
	tgr := New([]string{"env=prod"})
	a := tgr.Tags()
	a["env"] = "mutated"
	b := tgr.Tags()
	if b["env"] != "prod" {
		t.Error("Tags() should return an independent copy")
	}
}
