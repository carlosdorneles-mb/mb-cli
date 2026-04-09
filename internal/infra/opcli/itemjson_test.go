package opcli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUpsertConcealedFieldInItemJSON_NewSectionAndField(t *testing.T) {
	raw := []byte(`{
  "id": "item-1",
  "title": "mb-cli env / default",
  "vault": {"id": "v1", "name": "Private"},
  "category": "PASSWORD",
  "fields": [
    {"id": "username", "type": "STRING", "label": "username", "value": "u"}
  ]
}`)
	out, err := upsertConcealedFieldInItemJSON(raw, "API_KEY", "secret-val")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "API_KEY") || !strings.Contains(string(out), "secret-val") {
		t.Fatalf("missing field: %s", out)
	}
	if !strings.Contains(string(out), sectionMBCLI) {
		t.Fatalf("missing section: %s", out)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, hasVault := parsed["vault"]; hasVault {
		t.Fatal("edit payload must omit vault for op item edit stdin")
	}
}

func TestUpsertConcealedFieldInItemJSON_UpdateExisting(t *testing.T) {
	raw := []byte(`{
  "id": "item-1",
  "title": "t",
  "vault": {"id": "v1", "name": "Private"},
  "category": "PASSWORD",
  "sections": [{"id": "S1", "label": "MB CLI"}],
  "fields": [
    {
      "type": "CONCEALED",
      "label": "K",
      "value": "old",
      "section": {"id": "S1", "label": "MB CLI"}
    }
  ]
}`)
	out, err := upsertConcealedFieldInItemJSON(raw, "K", "new")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	fields := m["fields"].([]interface{})
	f0 := fields[0].(map[string]interface{})
	if f0["value"] != "new" {
		t.Fatalf("value=%v", f0["value"])
	}
	if _, hasVault := m["vault"]; hasVault {
		t.Fatal("edit payload must omit vault for op item edit stdin")
	}
}

func TestRemoveConcealedFieldFromItemJSON(t *testing.T) {
	raw := []byte(`{
  "id": "item-1",
  "title": "t",
  "vault": {"id": "v1", "name": "Private"},
  "category": "PASSWORD",
  "sections": [{"id": "S1", "label": "MB CLI"}],
  "fields": [
    {
      "type": "CONCEALED",
      "label": "K",
      "value": "x",
      "section": {"id": "S1", "label": "MB CLI"}
    },
    {"type": "STRING", "label": "other", "value": "y"}
  ]
}`)
	out, err := removeConcealedFieldFromItemJSON(raw, "K")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), `"label":"K"`) {
		t.Fatalf("field still present: %s", out)
	}
	if !strings.Contains(string(out), `"label":"other"`) {
		t.Fatalf("other field removed: %s", out)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, hasVault := parsed["vault"]; hasVault {
		t.Fatal("edit payload must omit vault for op item edit stdin")
	}
}

func TestFieldReferenceFromItemJSON_UsesReference(t *testing.T) {
	raw := []byte(`{
  "id": "item-uuid",
  "vault": {"id": "v", "name": "Private"},
  "fields": [
    {
      "label": "K",
      "reference": "op://Private/item-uuid/K",
      "section": {"label": "MB CLI"}
    }
  ]
}`)
	ref, err := fieldReferenceFromItemJSON(raw, "K")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "op://Private/item-uuid/K" {
		t.Fatalf("ref=%q", ref)
	}
}

func TestFieldReferenceFromItemJSON_Fallback(t *testing.T) {
	raw := []byte(`{
  "id": "item-uuid",
  "vault": {"name": "Work"},
  "fields": [
    {
      "label": "TOKEN",
      "type": "CONCEALED",
      "value": "x",
      "section": {"label": "MB CLI"}
    }
  ]
}`)
	ref, err := fieldReferenceFromItemJSON(raw, "TOKEN")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "op://Work/item-uuid/TOKEN" {
		t.Fatalf("ref=%q", ref)
	}
}
