package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NoteItem represents a Bitwarden note item structure
type NoteItem struct {
	Type       int        `json:"type"`
	Name       string     `json:"name"`
	Notes      string     `json:"notes"`
	FolderID   string     `json:"folderId"`
	SecureNote SecureNote `json:"secureNote"`
}

// SecureNote represents the secure note type
type SecureNote struct {
	Type int `json:"type"` // 0 = Text
}

// FullItem represents a full Bitwarden item (for getting item details)
type FullItem struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       int        `json:"type"`
	Notes      string     `json:"notes"`
	FolderID   string     `json:"folderId"`
	SecureNote SecureNote `json:"secureNote"`
}

// GetItemByName finds an item by name in the specified folder
func GetItemByName(folderID, itemName string) (*FullItem, error) {
	if err := requireBwInstalled(); err != nil {
		return nil, err
	}

	StartSpinner("Syncing...")

	syncRes := runBw("bw", []string{"sync"}, bwRunOptions{})
	if syncRes.Err != nil {
		syncErrMsg := strings.TrimSpace(string(syncRes.Output))
		if syncErrMsg != "" && !strings.Contains(syncErrMsg, "already synced") {
			StopSpinner()
			fmt.Printf("[INFO] Sync warning: %s (continuing anyway)\n", syncErrMsg)
		}
	}

	UpdateSpinnerMessage("Fetching items...")

	res := runBw("bw", []string{"list", "items", "--folderid", folderID}, bwRunOptions{})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to list items: %s", errorMsg)
	}

	var items []FullItem
	outputStr := strings.TrimSpace(string(res.Output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw list items command")
	}

	if isBwAuthRequiredMessage(outputStr) {
		return nil, ErrBitwardenLocked
	}

	if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
		return nil, fmt.Errorf("unexpected output from bw list items (not JSON): %s", outputStr)
	}

	if err := json.Unmarshal([]byte(outputStr), &items); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON (output: %s): %w", outputStr, err)
	}

	for _, item := range items {
		if item.Name == itemName {
			StopSpinner()
			return GetItemByID(item.ID)
		}
	}

	StopSpinner()
	return nil, nil
}

// GetItemByID retrieves a full item by ID using bw get item
func GetItemByID(itemID string) (*FullItem, error) {
	if err := requireBwInstalled(); err != nil {
		return nil, err
	}

	StartSpinner("Getting item...")
	defer StopSpinner()

	res := runBw("bw", []string{"get", "item", itemID}, bwRunOptions{})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to get item: %s", errorMsg)
	}

	var item FullItem
	outputStr := strings.TrimSpace(string(res.Output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw get item command")
	}

	if isBwAuthRequiredMessage(outputStr) {
		return nil, ErrBitwardenLocked
	}

	if !strings.HasPrefix(outputStr, "{") {
		return nil, fmt.Errorf("unexpected output from bw get item (not JSON): %s", outputStr)
	}

	if err := json.Unmarshal([]byte(outputStr), &item); err != nil {
		return nil, fmt.Errorf("failed to parse item JSON (output: %s): %w", outputStr, err)
	}

	return &item, nil
}

// CreateNoteItem creates a new note item in Bitwarden
func CreateNoteItem(folderID, name, notes string) error {
	if err := requireBwInstalled(); err != nil {
		return err
	}

	StartSpinner("Creating item...")
	defer StopSpinner()

	item := NoteItem{
		Type:     2,
		Name:     name,
		Notes:    notes,
		FolderID: folderID,
		SecureNote: SecureNote{
			Type: 0,
		},
	}

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item to JSON: %w", err)
	}

	templateRes := runBw("bw", []string{"get", "template", "item"}, bwRunOptions{})
	if templateRes.Err != nil {
		return createItemDirectly(itemJSON)
	}

	var template map[string]interface{}
	if err := json.Unmarshal(templateRes.Output, &template); err != nil {
		return createItemDirectly(itemJSON)
	}

	template["type"] = 2
	template["name"] = name
	template["notes"] = notes
	template["folderId"] = folderID
	template["secureNote"] = map[string]interface{}{
		"type": 0,
	}

	modifiedJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal modified template: %w", err)
	}

	return createItemWithEncode(modifiedJSON)
}

func createItemDirectly(itemJSON []byte) error {
	res := runBw("bw", []string{"create", "item"}, bwRunOptions{Stdin: itemJSON})
	if res.Err != nil {
		errorMsg := strings.TrimSpace(string(res.Output))
		if errorMsg == "" {
			errorMsg = res.Err.Error()
		}
		return fmt.Errorf("failed to create item: %s", errorMsg)
	}
	return nil
}

func createItemWithEncode(itemJSON []byte) error {
	encodeRes := runBw("bw", []string{"encode"}, bwRunOptions{Stdin: itemJSON, StdoutOnly: true})
	if encodeRes.Err != nil {
		return createItemDirectly(itemJSON)
	}

	createRes := runBw("bw", []string{"create", "item", strings.TrimSpace(string(encodeRes.Output))}, bwRunOptions{})
	if createRes.Err != nil {
		errorMsg := strings.TrimSpace(string(createRes.Output))
		if errorMsg == "" {
			errorMsg = createRes.Err.Error()
		}
		return fmt.Errorf("failed to create item: %s", errorMsg)
	}
	return nil
}

// UpdateNoteItem updates an existing note item's notes field
func UpdateNoteItem(itemID, notes string) error {
	if err := requireBwInstalled(); err != nil {
		return err
	}

	StartSpinner("Updating item...")
	defer StopSpinner()

	getRes := runBw("bw", []string{"get", "item", itemID}, bwRunOptions{})
	if getRes.Err != nil {
		errorMsg := strings.TrimSpace(string(getRes.Output))
		if errorMsg == "" {
			errorMsg = getRes.Err.Error()
		}
		return fmt.Errorf("failed to get item: %s", errorMsg)
	}

	var item map[string]interface{}
	outputStr := strings.TrimSpace(string(getRes.Output))
	if outputStr == "" {
		return fmt.Errorf("no output from bw get item command")
	}

	if err := json.Unmarshal([]byte(outputStr), &item); err != nil {
		return fmt.Errorf("failed to parse item JSON: %w", err)
	}

	item["notes"] = notes

	updatedJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal updated item: %w", err)
	}

	encodeRes := runBw("bw", []string{"encode"}, bwRunOptions{Stdin: updatedJSON, StdoutOnly: true})
	if encodeRes.Err != nil {
		editRes := runBw("bw", []string{"edit", "item", itemID}, bwRunOptions{Stdin: updatedJSON})
		if editRes.Err != nil {
			errorMsg := strings.TrimSpace(string(editRes.Output))
			if errorMsg == "" {
				errorMsg = editRes.Err.Error()
			}
			return fmt.Errorf("failed to update item: %s", errorMsg)
		}
		return nil
	}

	encodedStr := strings.TrimSpace(string(encodeRes.Output))
	editRes := runBw("bw", []string{"edit", "item", itemID, encodedStr}, bwRunOptions{})
	if editRes.Err != nil {
		errorMsg := strings.TrimSpace(string(editRes.Output))
		if errorMsg == "" {
			errorMsg = editRes.Err.Error()
		}
		return fmt.Errorf("failed to update item: %s", errorMsg)
	}

	return nil
}
