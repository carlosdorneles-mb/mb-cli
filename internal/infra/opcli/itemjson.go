package opcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const sectionMBCLI = "MB CLI"

// upsertConcealedFieldInItemJSON adds or updates a concealed field under sectionMBCLI.
func upsertConcealedFieldInItemJSON(itemJSON []byte, fieldLabel, value string) ([]byte, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(itemJSON, &root); err != nil {
		return nil, fmt.Errorf("item json: %w", err)
	}
	stripReadOnlyItemKeys(root)

	sectionID, sections := ensureSection(root, sectionMBCLI)
	root["sections"] = sections

	fields, ok := root["fields"].([]interface{})
	if !ok || fields == nil {
		fields = []interface{}{}
	}

	found := false
	newFields := make([]interface{}, 0, len(fields)+1)
	for _, f := range fields {
		fm, ok := f.(map[string]interface{})
		if !ok {
			newFields = append(newFields, f)
			continue
		}
		secLbl := sectionLabelFromField(fm)
		lbl, _ := fm["label"].(string)
		if secLbl == sectionMBCLI && lbl == fieldLabel {
			fm["type"] = "CONCEALED"
			fm["value"] = value
			if fm["section"] == nil {
				fm["section"] = map[string]interface{}{"id": sectionID, "label": sectionMBCLI}
			}
			found = true
		}
		newFields = append(newFields, fm)
	}
	if !found {
		newFields = append(newFields, map[string]interface{}{
			"type":  "CONCEALED",
			"label": fieldLabel,
			"value": value,
			"section": map[string]interface{}{
				"id":    sectionID,
				"label": sectionMBCLI,
			},
		})
	}
	root["fields"] = newFields
	return marshalItemJSONForOPEdit(root)
}

// removeConcealedFieldFromItemJSON removes a custom field under sectionMBCLI with the given label.
func removeConcealedFieldFromItemJSON(itemJSON []byte, fieldLabel string) ([]byte, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(itemJSON, &root); err != nil {
		return nil, fmt.Errorf("item json: %w", err)
	}
	stripReadOnlyItemKeys(root)

	fields, ok := root["fields"].([]interface{})
	if !ok || fields == nil {
		return marshalItemJSONForOPEdit(root)
	}

	newFields := make([]interface{}, 0, len(fields))
	for _, f := range fields {
		fm, ok := f.(map[string]interface{})
		if !ok {
			newFields = append(newFields, f)
			continue
		}
		secLbl := sectionLabelFromField(fm)
		lbl, _ := fm["label"].(string)
		if secLbl == sectionMBCLI && lbl == fieldLabel {
			continue
		}
		newFields = append(newFields, fm)
	}
	root["fields"] = newFields
	pruneEmptyMBCLISection(root)
	return marshalItemJSONForOPEdit(root)
}

// marshalItemJSONForOPEdit marshals item JSON for `op item edit` stdin.
// The 1Password CLI compares vault in the template to the item being edited; a mismatch
// (e.g. default "Private" in JSON vs item in "Employee") fails with identity inconsistencies.
// The item is already identified by the `op item edit <name>` argument, so vault may be omitted.
func marshalItemJSONForOPEdit(root map[string]interface{}) ([]byte, error) {
	delete(root, "vault")
	return json.Marshal(root)
}

func sectionLabelFromField(fm map[string]interface{}) string {
	sec, ok := fm["section"].(map[string]interface{})
	if !ok || sec == nil {
		return ""
	}
	lbl, _ := sec["label"].(string)
	return lbl
}

func ensureSection(
	root map[string]interface{},
	label string,
) (sectionID string, sections []interface{}) {
	sections, _ = root["sections"].([]interface{})
	if sections == nil {
		sections = []interface{}{}
	}
	for _, s := range sections {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		l, _ := sm["label"].(string)
		if l == label {
			sid, _ := sm["id"].(string)
			return sid, sections
		}
	}
	sectionID = "Section_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	sections = append(sections, map[string]interface{}{
		"id":    sectionID,
		"label": label,
	})
	return sectionID, sections
}

func pruneEmptyMBCLISection(root map[string]interface{}) {
	fields, _ := root["fields"].([]interface{})
	hasMBCLI := false
	for _, f := range fields {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		if sectionLabelFromField(fm) == sectionMBCLI {
			hasMBCLI = true
			break
		}
	}
	if hasMBCLI {
		return
	}
	sections, _ := root["sections"].([]interface{})
	if len(sections) == 0 {
		return
	}
	out := make([]interface{}, 0, len(sections))
	for _, s := range sections {
		sm, ok := s.(map[string]interface{})
		if !ok {
			out = append(out, s)
			continue
		}
		l, _ := sm["label"].(string)
		if l == sectionMBCLI {
			continue
		}
		out = append(out, s)
	}
	root["sections"] = out
}

func stripReadOnlyItemKeys(root map[string]interface{}) {
	for _, k := range []string{
		"created_at", "updated_at", "last_edited_by", "first_created_at",
		"additional_information", "files", "document", "open_totp",
	} {
		delete(root, k)
	}
}

// fieldReferenceFromItemJSON returns the op:// reference for a field label under sectionMBCLI.
func fieldReferenceFromItemJSON(itemJSON []byte, fieldLabel string) (string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(itemJSON, &root); err != nil {
		return "", err
	}
	fields, _ := root["fields"].([]interface{})
	for _, f := range fields {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		if sectionLabelFromField(fm) != sectionMBCLI {
			continue
		}
		lbl, _ := fm["label"].(string)
		if lbl != fieldLabel {
			continue
		}
		if ref, ok := fm["reference"].(string); ok && ref != "" {
			return ref, nil
		}
		return buildFallbackReference(root, fieldLabel)
	}
	return "", fmt.Errorf("campo %q não encontrado na secção %s", fieldLabel, sectionMBCLI)
}

func buildFallbackReference(root map[string]interface{}, fieldLabel string) (string, error) {
	vaultName, vaultID := vaultNameOrID(root)
	itemID, _ := root["id"].(string)
	if vaultName == "" && vaultID == "" {
		return "", fmt.Errorf("item sem vault na resposta do op")
	}
	if itemID == "" {
		return "", fmt.Errorf("item sem id na resposta do op")
	}
	v := vaultName
	if v == "" {
		v = vaultID
	}
	return fmt.Sprintf("op://%s/%s/%s", v, itemID, fieldLabel), nil
}

func vaultNameOrID(root map[string]interface{}) (name, id string) {
	v, ok := root["vault"].(map[string]interface{})
	if !ok || v == nil {
		return "", ""
	}
	name, _ = v["name"].(string)
	id, _ = v["id"].(string)
	return name, id
}
