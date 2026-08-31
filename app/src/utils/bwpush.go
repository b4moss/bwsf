package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return nil, fmt.Errorf("bw command is not installed")
	}

	// Start spinner for sync
	StartSpinner("Syncing...")

	// Sync with server first to ensure we have the latest data
	syncCmd := exec.Command("bw", "sync")
	syncOutput, syncErr := syncCmd.CombinedOutput()
	if syncErr != nil {
		// Sync failure is not critical, but log it
		syncErrMsg := strings.TrimSpace(string(syncOutput))
		if syncErrMsg != "" && !strings.Contains(syncErrMsg, "already synced") {
			// Only log if it's not just "already synced" message
			StopSpinner()
			fmt.Printf("[INFO] Sync warning: %s (continuing anyway)\n", syncErrMsg)
		}
	}

	// Update spinner message for fetching items
	UpdateSpinnerMessage("Fetching items...")

	// Execute bw list items command with folder filter
	cmd := exec.Command("bw", "list", "items", "--folderid", folderID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to list items: %s", errorMsg)
	}

	// Parse JSON output
	var items []FullItem
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw list items command")
	}

	// Check if Bitwarden CLI requires auth (locked / unauthenticated)
	if isBwAuthRequiredMessage(outputStr) {
		return nil, ErrBitwardenLocked
	}

	// Check if output looks like JSON
	if !strings.HasPrefix(outputStr, "[") && !strings.HasPrefix(outputStr, "{") {
		return nil, fmt.Errorf("unexpected output from bw list items (not JSON): %s", outputStr)
	}

	if err := json.Unmarshal([]byte(outputStr), &items); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON (output: %s): %w", outputStr, err)
	}

	// Find item by name
	for _, item := range items {
		if item.Name == itemName {
			// Stop spinner before calling GetItemByID (it has its own spinner)
			StopSpinner()
			// Get full item details using bw get item to ensure we have the latest data
			return GetItemByID(item.ID)
		}
	}

	StopSpinner()
	return nil, nil // Item not found, but no error
}

// GetItemByID retrieves a full item by ID using bw get item
func GetItemByID(itemID string) (*FullItem, error) {
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return nil, fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner("Getting item...")
	defer StopSpinner()

	// Execute bw get item command
	cmd := exec.Command("bw", "get", "item", itemID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		if isBwAuthRequiredMessage(errorMsg) {
			return nil, ErrBitwardenLocked
		}
		return nil, fmt.Errorf("failed to get item: %s", errorMsg)
	}

	// Parse JSON output
	var item FullItem
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("no output from bw get item command")
	}

	// Check if Bitwarden CLI requires auth (locked / unauthenticated)
	if isBwAuthRequiredMessage(outputStr) {
		return nil, ErrBitwardenLocked
	}

	// Check if output looks like JSON
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
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner("Creating item...")
	defer StopSpinner()

	// Create note item structure
	item := NoteItem{
		Type:     2, // Secure Note type
		Name:     name,
		Notes:    notes,
		FolderID: folderID,
		SecureNote: SecureNote{
			Type: 0, // Text type
		},
	}

	// Marshal to JSON
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item to JSON: %w", err)
	}

	// Get template and modify it
	// First, get the template
	templateCmd := exec.Command("bw", "get", "template", "item")
	templateOutput, err := templateCmd.CombinedOutput()
	if err != nil {
		// If template command fails, try direct creation
		return createItemDirectly(itemJSON)
	}

	// Parse template and modify
	var template map[string]interface{}
	if err := json.Unmarshal(templateOutput, &template); err != nil {
		// If parsing fails, try direct creation
		return createItemDirectly(itemJSON)
	}

	// Update template with our data
	template["type"] = 2
	template["name"] = name
	template["notes"] = notes
	template["folderId"] = folderID
	template["secureNote"] = map[string]interface{}{
		"type": 0,
	}

	// Marshal modified template
	modifiedJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal modified template: %w", err)
	}

	// Encode and create item
	return createItemWithEncode(modifiedJSON)
}

// createItemDirectly creates item by piping JSON directly
func createItemDirectly(itemJSON []byte) error {
	cmd := exec.Command("bw", "create", "item")
	cmd.Stdin = bytes.NewReader(itemJSON)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("failed to create item: %s", errorMsg)
	}
	return nil
}

// createItemWithEncode creates item using encode command
func createItemWithEncode(itemJSON []byte) error {
	// Encode the JSON
	encodeCmd := exec.Command("bw", "encode")
	encodeCmd.Stdin = bytes.NewReader(itemJSON)
	encodedOutput, err := encodeCmd.Output()
	if err != nil {
		// If encode fails, try direct creation
		return createItemDirectly(itemJSON)
	}

	// Create item with encoded data
	createCmd := exec.Command("bw", "create", "item", strings.TrimSpace(string(encodedOutput)))
	output, err := createCmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("failed to create item: %s", errorMsg)
	}
	return nil
}

// UpdateNoteItem updates an existing note item's notes field
func UpdateNoteItem(itemID, notes string) error {
	// Check if bw command exists
	_, err := exec.LookPath("bw")
	if err != nil {
		return fmt.Errorf("bw command is not installed")
	}

	// Start spinner
	StartSpinner("Updating item...")
	defer StopSpinner()

	// Get the existing item
	getCmd := exec.Command("bw", "get", "item", itemID)
	output, err := getCmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(output))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("failed to get item: %s", errorMsg)
	}

	// Parse existing item
	var item map[string]interface{}
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return fmt.Errorf("no output from bw get item command")
	}

	if err := json.Unmarshal([]byte(outputStr), &item); err != nil {
		return fmt.Errorf("failed to parse item JSON: %w", err)
	}

	// Update notes field
	item["notes"] = notes

	// Marshal updated item
	updatedJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal updated item: %w", err)
	}

	// Encode the JSON using bw encode
	encodeCmd := exec.Command("bw", "encode")
	encodeCmd.Stdin = bytes.NewReader(updatedJSON)
	encodedOutput, err := encodeCmd.Output()
	if err != nil {
		// If encode fails, try direct edit (for compatibility)
		editCmd := exec.Command("bw", "edit", "item", itemID)
		editCmd.Stdin = bytes.NewReader(updatedJSON)
		editOutput, err := editCmd.CombinedOutput()
		if err != nil {
			errorMsg := strings.TrimSpace(string(editOutput))
			if errorMsg == "" {
				errorMsg = err.Error()
			}
			return fmt.Errorf("failed to update item: %s", errorMsg)
		}
		return nil
	}

	// Edit the item with encoded data
	encodedStr := strings.TrimSpace(string(encodedOutput))
	editCmd := exec.Command("bw", "edit", "item", itemID, encodedStr)
	editOutput, err := editCmd.CombinedOutput()
	if err != nil {
		errorMsg := strings.TrimSpace(string(editOutput))
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return fmt.Errorf("failed to update item: %s", errorMsg)
	}

	return nil
}
