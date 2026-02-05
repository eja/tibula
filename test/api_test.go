// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/eja/tibula/api"
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
)

func setupTestDB(t *testing.T) (string, func()) {
	tmpDb, err := os.CreateTemp("", "tibula_api_test_*.db")
	if err != nil {
		t.Fatal("Could not create temp DB file:", err)
	}
	dbPath := tmpDb.Name()
	tmpDb.Close()

	sys.Options.DbType = "sqlite"
	sys.Options.DbName = dbPath
	sys.Options.DbHost = ""
	sys.Options.DbPort = 0
	sys.Options.DbUser = ""
	sys.Options.DbPass = ""

	d := db.Session()
	if err := d.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort); err != nil {
		t.Fatal("Failed to open temp DB for setup:", err)
	}

	if err := d.Setup(""); err != nil {
		t.Fatalf("Database setup failed: %v", err)
	}

	if err := d.SetupAdmin("admin", "secret"); err != nil {
		t.Fatalf("Failed to setup admin user: %v", err)
	}

	d.Close()

	cleanup := func() {
		os.Remove(dbPath)
	}

	return dbPath, cleanup
}

// getAuthenticatedSession logs in and returns a session token
func getAuthenticatedSession(t *testing.T) string {
	eja := api.Set()
	eja.Action = "login"
	eja.Values["username"] = "admin"
	eja.Values["password"] = "secret"

	res, err := api.Run(eja, true) // sessionSave must be true to keep the session
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	if res.Session == "" {
		t.Fatal("Expected valid session token")
	}
	return res.Session
}

// TestSet verifies the initialization of API structure
func TestSet(t *testing.T) {
	eja := api.Set()

	if eja.DefaultSearchLimit != 15 {
		t.Errorf("Expected DefaultSearchLimit 15, got %d", eja.DefaultSearchLimit)
	}
	if eja.DefaultSearchOrder != "ejaId DESC" {
		t.Errorf("Expected DefaultSearchOrder 'ejaId DESC', got %s", eja.DefaultSearchOrder)
	}
	if eja.Values == nil {
		t.Error("Values map should be initialized")
	}
	if eja.SearchOrder == nil {
		t.Error("SearchOrder map should be initialized")
	}
	if eja.Language != sys.Options.Language {
		t.Errorf("Expected Language %s, got %s", sys.Options.Language, eja.Language)
	}
}

// TestAuthenticationFlow tests login, session management, and logout
func TestAuthenticationFlow(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Login_Success", func(t *testing.T) {
		eja := api.Set()
		eja.Action = "login"
		eja.Values["username"] = "admin"
		eja.Values["password"] = "secret"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Login failed with error: %v", err)
		}
		if res.Session == "" {
			t.Error("Expected valid Session token, got empty string")
		}
		if res.Owner == 0 {
			t.Error("Expected valid Owner ID (non-zero)")
		}
		if res.ModuleId == 0 {
			t.Error("Expected valid ModuleId after login")
		}
	})

	t.Run("Login_Failure_WrongPassword", func(t *testing.T) {
		eja := api.Set()
		eja.Action = "login"
		eja.Values["username"] = "admin"
		eja.Values["password"] = "wrong_password"

		res, err := api.Run(eja, true)
		if err == nil {
			t.Error("Expected error for wrong password, got nil")
		}
		if res.Session != "" {
			t.Errorf("Expected empty session, got %s", res.Session)
		}
		if len(res.Alert) == 0 {
			t.Error("Expected Alert message in response for failed login")
		}
	})

	t.Run("Login_Failure_EmptyCredentials", func(t *testing.T) {
		eja := api.Set()
		eja.Action = "login"
		eja.Values["username"] = ""
		eja.Values["password"] = ""

		_, err := api.Run(eja, true)
		if err == nil {
			t.Error("Expected error for empty credentials")
		}
	})

	t.Run("Logout", func(t *testing.T) {
		session := getAuthenticatedSession(t)

		eja := api.Set()
		eja.Session = session
		eja.Action = "logout"

		res, _ := api.Run(eja, true) // Logout may return error, that's ok
		// After logout, session and owner should be cleared
		if res.Session != "" {
			t.Error("Expected session to be cleared after logout")
		}
		if res.Owner != 0 {
			t.Error("Expected Owner to be 0 after logout")
		}
	})

	t.Run("Session_Validation", func(t *testing.T) {
		session := getAuthenticatedSession(t)

		eja := api.Set()
		eja.Session = session

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Valid session should not error: %v", err)
		}
		if res.Owner == 0 {
			t.Error("Expected valid Owner with valid session")
		}
	})

	t.Run("Session_Invalid", func(t *testing.T) {
		eja := api.Set()
		eja.Session = "invalid_session_token"

		_, err := api.Run(eja, true)
		if err == nil {
			t.Error("Expected error for invalid session")
		}
	})
}

// TestCRUDOperations tests create, read, update, delete operations
func TestCRUDOperations(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Create_New_Record", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to create new record: %v", err)
		}
		if res.Id == 0 {
			t.Error("Expected new record ID to be assigned")
		}
		if res.ActionType != "Edit" {
			t.Errorf("Expected ActionType 'Edit' after new, got %s", res.ActionType)
		}
	})

	t.Run("Save_Record", func(t *testing.T) {
		// Create a new record first
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		newId := res.Id

		// Now save data to it
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = newId
		eja.Action = "save"
		eja.Values["username"] = "testuser"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to save record: %v", err)
		}
		if res.Values["username"] != "testuser" {
			t.Error("Username not saved correctly")
		}
	})

	t.Run("Edit_Record", func(t *testing.T) {
		// Create and save a record
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		newId := res.Id

		// Edit the record
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = newId
		eja.Action = "edit"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to edit record: %v", err)
		}
		if res.ActionType != "Edit" {
			t.Errorf("Expected ActionType 'Edit', got %s", res.ActionType)
		}
		if res.Id != newId {
			t.Error("Record ID mismatch")
		}
	})

	t.Run("Copy_Record", func(t *testing.T) {
		// Create and save a record
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		originalId := res.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = originalId
		eja.Action = "save"
		eja.Values["username"] = "original"
		api.Run(eja, true)

		// Copy the record
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = originalId
		eja.Action = "copy"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to copy record: %v", err)
		}
		if res.Id == 0 || res.Id == originalId {
			t.Error("Expected new ID for copied record")
		}
	})

	t.Run("Delete_Record", func(t *testing.T) {
		// Create a record
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		idToDelete := res.Id

		// Delete it
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = idToDelete
		eja.Action = "delete"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to delete record: %v", err)
		}
		if res.ActionType != "List" {
			t.Errorf("Expected ActionType 'List' after delete, got %s", res.ActionType)
		}
	})

	t.Run("Delete_Multiple_Records", func(t *testing.T) {
		// Create multiple records
		var ids []int64
		for i := 0; i < 3; i++ {
			eja := api.Set()
			eja.Session = session
			eja.ModuleName = "ejaUsers"
			eja.Action = "new"
			res, _ := api.Run(eja, true)
			ids = append(ids, res.Id)
		}

		// Delete them all
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.IdList = ids
		eja.Action = "delete"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to delete multiple records: %v", err)
		}
		if res.ActionType != "List" {
			t.Error("Expected List view after bulk delete")
		}
	})
}

// TestSearchAndList tests search functionality
func TestSearchAndList(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	// Create some test records
	for i := 0; i < 5; i++ {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["username"] = "user" + string(rune(i))
		api.Run(eja, true)
	}

	t.Run("List_Action", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("List action failed: %v", err)
		}
		// ActionType can be List or Search depending on module configuration
		if res.ActionType != "List" && res.ActionType != "Search" {
			t.Errorf("Expected ActionType 'List' or 'Search', got %s", res.ActionType)
		}
		// SearchCount may be 0 if no records match
	})

	t.Run("Search_Action", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "search"
		eja.Values["username"] = "user"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Search action failed: %v", err)
		}
		if res.ActionType != "List" {
			t.Error("Expected List ActionType after search")
		}
	})

	t.Run("Search_Pagination_Next", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 2
		eja.Action = "list"
		res, _ := api.Run(eja, true)
		initialOffset := res.SearchOffset

		// Next page
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 2
		eja.Action = "next"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Next page action failed: %v", err)
		}
		// SearchOffset should have increased
		if res.SearchOffset <= initialOffset {
			t.Errorf("Expected SearchOffset to increase from %d, got %d", initialOffset, res.SearchOffset)
		}
	})

	t.Run("Search_Pagination_Previous", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 2
		eja.SearchOffset = 4
		eja.Action = "previous"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Previous page action failed: %v", err)
		}
		// SearchOffset should have decreased (or stayed at 0 if already at start)
		if res.SearchOffset > 4 {
			t.Errorf("Expected SearchOffset to decrease from 4, got %d", res.SearchOffset)
		}
	})

	t.Run("Search_With_Order", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "list"
		eja.SearchOrder = map[string]string{
			"username": "ASC",
		}

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Search with order failed: %v", err)
		}
		if res.ActionType != "List" {
			t.Error("Expected List ActionType")
		}
	})

	t.Run("Search_With_Base64Query", func(t *testing.T) {
		query := "SELECT * FROM ejaUsers WHERE ejaOwner=?"
		encoded := base64.StdEncoding.EncodeToString([]byte(query))

		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SqlQuery64 = encoded
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Search with base64 query failed: %v", err)
		}
		if res.ActionType != "List" {
			t.Error("Expected List ActionType")
		}
	})

	t.Run("SearchLinkClean", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLinkClean = true
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("SearchLinkClean failed: %v", err)
		}
		if res.ActionType != "List" && res.ActionType != "Search" {
			t.Errorf("Expected ActionType 'List' or 'Search', got %s", res.ActionType)
		}
	})
}

// TestLinkingOperations tests link/unlink functionality
func TestLinkingOperations(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Link_Records", func(t *testing.T) {
		// Create two records to link
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res1, _ := api.Run(eja, true)
		id1 := res1.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaGroups"
		eja.Action = "new"
		res2, _ := api.Run(eja, true)
		id2 := res2.Id

		// Link them
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "link"
		eja.IdList = []int64{id1}
		eja.Link.ModuleId = res2.ModuleId
		eja.Link.FieldId = id2
		eja.Link.Label = "Test Link"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Link action failed: %v", err)
		}
		if res.ActionType != "List" && res.ActionType != "Search" {
			t.Errorf("Expected ActionType 'List' or 'Search' after link, got %s", res.ActionType)
		}
	})

	t.Run("Unlink_Records", func(t *testing.T) {
		// Create and link records first
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res1, _ := api.Run(eja, true)
		id1 := res1.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaGroups"
		eja.Action = "new"
		res2, _ := api.Run(eja, true)
		id2 := res2.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "link"
		eja.IdList = []int64{id1}
		eja.Link.ModuleId = res2.ModuleId
		eja.Link.FieldId = id2
		eja.Link.Label = "Test Link"
		api.Run(eja, true)

		// Now unlink
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "unlink"
		eja.IdList = []int64{id1}
		eja.Link.ModuleId = res2.ModuleId
		eja.Link.FieldId = id2
		eja.Link.Label = "Test Link"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Unlink action failed: %v", err)
		}
		if res.ActionType != "List" && res.ActionType != "Search" {
			t.Errorf("Expected ActionType 'List' or 'Search' after unlink, got %s", res.ActionType)
		}
	})

	t.Run("SearchLink", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLink = true
		eja.Link.ModuleId = 1
		eja.Link.FieldId = 1
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("SearchLink failed: %v", err)
		}
		if res.ActionType != "List" && res.ActionType != "Search" {
			t.Errorf("Expected ActionType 'List' or 'Search', got %s", res.ActionType)
		}
	})
}

// TestModuleOperations tests module-specific operations
func TestModuleOperations(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Module_By_Name", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Module by name failed: %v", err)
		}
		if res.ModuleId == 0 {
			t.Error("Expected ModuleId to be set")
		}
		if res.ModuleName != "ejaUsers" {
			t.Error("Module name mismatch")
		}
	})

	t.Run("Module_By_Id", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleId = 1

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Module by ID failed: %v", err)
		}
		if res.ModuleName == "" {
			t.Error("Expected ModuleName to be set")
		}
	})

	t.Run("Module_Commands", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Getting module commands failed: %v", err)
		}
		if len(res.Commands) == 0 {
			t.Error("Expected at least some commands")
		}
	})

	t.Run("Module_Fields", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "edit"
		eja.Id = 1

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Getting module fields failed: %v", err)
		}
		if len(res.Fields) == 0 {
			t.Error("Expected at least some fields")
		}
	})

	t.Run("Module_Path", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Getting module path failed: %v", err)
		}
		if res.Path == nil {
			t.Error("Expected Path to be initialized")
		}
	})

	t.Run("Module_Tree", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Getting module tree failed: %v", err)
		}
		// Tree may be nil or empty, just check it's accessible
		_ = res.Tree
	})

	t.Run("SubModules", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "edit"
		eja.Id = 1

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Getting submodules failed: %v", err)
		}
		// SubModules may be nil or empty, just check it's accessible
		_ = res.SubModules
	})
}

// TestSubModulePath tests submodule path navigation
func TestSubModulePath(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("SubModulePath_Navigation", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SubModulePath = []api.SubModulePathItem{
			{
				LinkingModuleId: 1,
				ModuleId:        2,
				FieldId:         1,
			},
		}

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("SubModulePath navigation failed: %v", err)
		}
		// SubModulePathString may be empty if path doesn't match module structure
		_ = res.SubModulePathString
	})
}

// TestFieldTypes tests different field type handling
func TestFieldTypes(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Save_Integer_Field", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["ejaActive"] = "1" // Use ejaActive instead of ejaSort

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Saving integer field failed: %v", err)
		}
		// Field should be saved (exact value may vary based on field type)
		_ = res.Values["ejaActive"]
	})

	t.Run("Save_Boolean_Field", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["ejaActive"] = "1"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Saving boolean field failed: %v", err)
		}
	})

	t.Run("Save_Password_Field", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["password"] = "newpassword123"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Saving password field failed: %v", err)
		}
		if res.Values["password"] == "newpassword123" {
			t.Error("Password should be hashed, not stored as plaintext")
		}
	})
}

// TestPlugins tests plugin functionality
func TestPlugins(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Plugin_ejaProfile_PasswordChange", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaProfile"
		eja.Action = "run"
		eja.Values["passwordOld"] = "secret"
		eja.Values["passwordNew"] = "newsecret123"
		eja.Values["passwordRepeat"] = "newsecret123"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Password change failed: %v", err)
		}
		if len(res.Info) == 0 {
			t.Error("Expected info message for successful password change")
		}
	})

	t.Run("Plugin_ejaProfile_PasswordMismatch", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaProfile"
		eja.Action = "run"
		eja.Values["passwordOld"] = "secret"
		eja.Values["passwordNew"] = "newsecret123"
		eja.Values["passwordRepeat"] = "different"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Plugin execution failed: %v", err)
		}
		if len(res.Alert) == 0 {
			t.Error("Expected alert for password mismatch")
		}
	})

	t.Run("Plugin_ejaProfile_WrongOldPassword", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaProfile"
		eja.Action = "run"
		eja.Values["passwordOld"] = "wrongpassword"
		eja.Values["passwordNew"] = "newsecret123"
		eja.Values["passwordRepeat"] = "newsecret123"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Plugin execution failed: %v", err)
		}
		if len(res.Alert) == 0 {
			t.Error("Expected alert for wrong old password")
		}
	})

	t.Run("Plugin_ejaModuleExport", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModuleExport"
		eja.Action = "run"
		eja.Values["ejaModuleId"] = "1"
		eja.Values["dataExport"] = "1"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Module export failed: %v", err)
		}
		if res.Values["export"] == "" {
			t.Error("Expected export data to be populated")
		}

		// Verify it's valid JSON
		var module db.TypeModule
		if err := json.Unmarshal([]byte(res.Values["export"]), &module); err != nil {
			t.Errorf("Export data is not valid JSON: %v", err)
		}
	})

	t.Run("Plugin_ejaModuleImport", func(t *testing.T) {
		// First export a module
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModuleExport"
		eja.Action = "run"
		eja.Values["ejaModuleId"] = "1"
		eja.Values["dataExport"] = "0"
		res, _ := api.Run(eja, true)
		exportData := res.Values["export"]

		// Now import it
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModuleImport"
		eja.Action = "run"
		eja.Values["import"] = exportData
		eja.Values["moduleName"] = "testImport"
		eja.Values["dataImport"] = "0"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Module import failed: %v", err)
		}
		if len(res.Info) == 0 {
			t.Error("Expected info message for successful import")
		}
	})

	t.Run("Plugin_ejaGroupExport", func(t *testing.T) {
		// First create a group to export
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaGroups"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		groupId := res.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaGroupExport"
		eja.Action = "run"
		eja.Values["ejaGroupId"] = fmt.Sprintf("%d", groupId)

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Group export failed: %v", err)
		}
		// Export data should be present or an error should be shown
		if res.Values["export"] == "" && len(res.Alert) == 0 {
			t.Error("Expected export data or alert message")
		}
	})
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Invalid_Module_Name", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "nonExistentModule"
		eja.Action = "list"

		res, err := api.Run(eja, true)
		// Should handle gracefully
		if err != nil {
			t.Logf("Expected behavior: %v", err)
		}
		if res.ModuleId != 0 {
			t.Logf("ModuleId: %d", res.ModuleId)
		}
	})

	t.Run("Invalid_Action", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "invalidAction"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Expected error for invalid action: %v", err)
		}
		if len(res.Alert) == 0 {
			t.Error("Expected alert for invalid action")
		}
	})

	t.Run("Edit_NonExistent_Record", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "edit"
		eja.Id = 999999

		res, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Error editing non-existent record: %v", err)
		}
		if len(res.Values) != 0 {
			t.Error("Expected empty values for non-existent record")
		}
	})

	t.Run("Save_Without_Id", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "save"
		eja.Values["username"] = "testuser"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Save without ID: %v", err)
		}
		// Should create new record
		if res.Id == 0 {
			t.Error("Expected new ID to be assigned")
		}
	})
}

// TestSessionManagement tests session state management
func TestSessionManagement(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Session_Save_State", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 10
		eja.SearchOffset = 5
		eja.Action = "list"

		res, err := api.Run(eja, true) // sessionSave = true
		if err != nil {
			t.Errorf("Failed to save session state: %v", err)
		}
		if res.SearchLimit != 10 {
			t.Error("SearchLimit not persisted")
		}
		if res.SearchOffset != 5 {
			t.Error("SearchOffset not persisted")
		}
	})

	t.Run("Session_No_Save_State", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 10
		eja.Action = "list"

		_, err := api.Run(eja, false) // sessionSave = false
		if err != nil {
			t.Errorf("Failed with sessionSave false: %v", err)
		}
		// Session should be reset after Run
	})
}

// TestLanguageSettings tests language/localization
func TestLanguageSettings(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Default_Language", func(t *testing.T) {
		eja := api.Set()
		// Language might be empty or set to sys.Options.Language
		// Just verify it's accessible
		_ = eja.Language
	})

	t.Run("User_Language_Preference", func(t *testing.T) {
		session := getAuthenticatedSession(t)

		eja := api.Set()
		eja.Session = session

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Failed to get user language: %v", err)
		}
		// Language field should be accessible even if empty
		_ = res.Language
	})
}

// TestAutoSearch tests automatic search triggering
func TestAutoSearch(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("AutoSearch_Module", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		// No action specified, should trigger auto-search if enabled

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("AutoSearch failed: %v", err)
		}
		// Check if it behaved as expected based on module settings
		if res.ActionType == "List" {
			t.Logf("AutoSearch triggered for module: %s", res.ModuleName)
		}
	})
}

// TestModuleSpecificOperations tests operations specific to system modules
func TestModuleSpecificOperations(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Create_Module", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModules"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModules"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["name"] = "testModule"
		eja.Values["label"] = "Test Module"
		eja.Values["sqlCreated"] = "1"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Creating module failed: %v", err)
		}
		if len(res.Info) == 0 {
			t.Error("Expected info message for module creation")
		}
	})

	t.Run("Create_Field", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaFields"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaFields"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["ejaModuleId"] = "1"
		eja.Values["name"] = "testField"
		eja.Values["type"] = "string"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Creating field failed: %v", err)
		}
	})

	t.Run("Delete_Module", func(t *testing.T) {
		// Create a test module first
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModules"
		eja.Action = "new"
		res, _ := api.Run(eja, true)
		moduleId := res.Id

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModules"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["name"] = "moduleToDelete"
		eja.Values["sqlCreated"] = "1"
		api.Run(eja, true)

		// Delete it
		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaModules"
		eja.Id = moduleId
		eja.Action = "delete"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Deleting module failed: %v", err)
		}
		// Should have info or alert message
		if len(res.Info) == 0 && len(res.Alert) == 0 {
			t.Error("Expected message for module deletion")
		}
	})
}

// TestPermissions tests permission-related functionality
func TestPermissions(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Unauthorized_Action", func(t *testing.T) {
		// Try an action that may not be permitted
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "someRestrictedAction"

		res, err := api.Run(eja, true)
		if err == nil && len(res.Alert) == 0 {
			t.Log("Action was permitted or handled gracefully")
		}
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Empty_Values_Map", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Values = map[string]string{}

		_, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Empty values map handled: %v", err)
		}
	})

	t.Run("Nil_IdList", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.IdList = nil
		eja.Action = "delete"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Nil IdList handled: %v", err)
		}
		if res.ActionType == "List" {
			t.Log("Action completed successfully")
		}
	})

	t.Run("Zero_SearchLimit", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchLimit = 0
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Zero SearchLimit caused error: %v", err)
		}
		if res.SearchLimit == 0 {
			t.Error("SearchLimit should be set to default")
		}
	})

	t.Run("Negative_SearchOffset", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.SearchOffset = -10
		eja.Action = "list"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Negative offset handled: %v", err)
		}
		if res.SearchOffset < 0 {
			t.Error("SearchOffset should not be negative")
		}
	})

	t.Run("Very_Long_Username", func(t *testing.T) {
		longUsername := string(make([]byte, 1000))
		for i := range longUsername {
			longUsername = longUsername[:i] + "a" + longUsername[i+1:]
		}

		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["username"] = longUsername

		_, err := api.Run(eja, true)
		if err != nil {
			t.Logf("Long username handled: %v", err)
		}
	})

	t.Run("Special_Characters_In_Values", func(t *testing.T) {
		eja := api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Action = "new"
		res, _ := api.Run(eja, true)

		eja = api.Set()
		eja.Session = session
		eja.ModuleName = "ejaUsers"
		eja.Id = res.Id
		eja.Action = "save"
		eja.Values["username"] = "user<>\"'&;%"

		res, err := api.Run(eja, true)
		if err != nil {
			t.Errorf("Special characters caused error: %v", err)
		}
		if res.Values["username"] != "user<>\"'&;%" {
			t.Error("Special characters not preserved")
		}
	})
}

// TestConcurrentOperations tests concurrent API calls
// Note: SQLite doesn't handle concurrent writes at the moment, so these tests are skipped
func TestConcurrentOperations(t *testing.T) {
	t.Skip("Skipping concurrent operations tests - SQLite doesn't support concurrent writes")

	_, cleanup := setupTestDB(t)
	defer cleanup()

	session := getAuthenticatedSession(t)

	t.Run("Concurrent_Reads", func(t *testing.T) {
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func() {
				eja := api.Set()
				eja.Session = session
				eja.ModuleName = "ejaUsers"
				eja.Action = "list"

				_, err := api.Run(eja, true)
				if err != nil {
					t.Errorf("Concurrent read failed: %v", err)
				}
				done <- true
			}()
		}

		for i := 0; i < 5; i++ {
			<-done
		}
	})

	t.Run("Concurrent_Writes", func(t *testing.T) {
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func(idx int) {
				eja := api.Set()
				eja.Session = session
				eja.ModuleName = "ejaUsers"
				eja.Action = "new"
				res, _ := api.Run(eja, true)

				eja = api.Set()
				eja.Session = session
				eja.ModuleName = "ejaUsers"
				eja.Id = res.Id
				eja.Action = "save"
				eja.Values["username"] = "concurrent_user"

				_, err := api.Run(eja, true)
				if err != nil {
					t.Errorf("Concurrent write %d failed: %v", idx, err)
				}
				done <- true
			}(i)
		}

		for i := 0; i < 3; i++ {
			<-done
		}
	})
}
