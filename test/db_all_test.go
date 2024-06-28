// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package test

import (
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
	"testing"
)

func TestModule(t *testing.T) {
	var tableId int64
	tableName := "ejaUsers"
	fieldName := "username"

	sys.Configure()

	db := db.Session()

	t.Run("Open db", func(t *testing.T) {
		if err := db.Open("sqlite", ":memory:", "", "", "", 0); err != nil {
			t.Error("Cannot open database:", err)
		}
	})

	t.Run("Populate db core", func(t *testing.T) {
		if err := db.Setup("../" + sys.Options.DbSetupPath); err != nil {
			t.Error("Setup error", err)
		}
		if err := db.SetupAdmin("test", "test"); err != nil {
			t.Error("Admin setup error", err)
		}
	})

	t.Run("Get Module Id", func(t *testing.T) {
		tableId = db.ModuleGetIdByName(tableName)
		if tableId < 1 {
			t.Error("Cannot retrieve module id")
		}
	})

	t.Run("Add first user", func(t *testing.T) {
		id, err := db.New(1, tableId)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Put(1, tableId, id, fieldName, "admin"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Add second user", func(t *testing.T) {
		id, err := db.New(1, tableId)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Put(1, tableId, id, fieldName, "manager"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Insert and put a third value", func(t *testing.T) {
		id, err := db.New(2, tableId)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Put(2, tableId, id, fieldName, "user"); err != nil {
			t.Fatal(err)
		}
		if err := db.Put(2, tableId, id, "defaultModuleId", "3"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Check the owner list for a specific module", func(t *testing.T) {
		list := db.Owners(1, tableId)
		if len(list) != 4 {
			t.Errorf("Expected list length to be 3, got %v", list)
		}
	})

	t.Run("Retrieve data for different owners", func(t *testing.T) {
		row1, err := db.Get(1, tableId, 2)
		if err != nil {
			t.Fatal(err)
		}
		row2, err := db.Get(1, tableId, 3)
		if err != nil {
			t.Fatal(err)
		}
		row3, err := db.Get(2, tableId, 4)
		if err != nil {
			t.Fatal(err)
		}
		if len(row1) < 1 || len(row2) < 1 || len(row3) < 1 {
			t.Errorf("Not all data have been returned")
		}
	})

	//? to check for actionType
	t.Run("Retrieve command list for module:ejaUsers userId:1 mode:List", func(t *testing.T) {
		values, err := db.Commands(1, db.ModuleGetIdByName(tableName), "Edit")
		if err != nil {
			t.Fatal(err)
		}
		if len(values) < 1 {
			t.Errorf("Commands not found for this module")
		}
	})

	//check module path and tree
	t.Run("Check module path and tree for ejaAdministration", func(t *testing.T) {
		tableId := db.ModuleGetIdByName("ejaAdministration")
		path := db.ModulePath(1, tableId)
		if len(path) != 1 {
			t.Errorf("Module path not valid: %v", path)
		} else {
			tree := db.ModuleTree(1, tableId, path)
			if len(tree) != 3 {
				t.Errorf("Module tree not valid: %v", tree)
			}
		}
	})

	//?
	t.Run("Check the field for module ejaUsers", func(t *testing.T) {
		values, err := db.Fields(0, db.ModuleGetIdByName(tableName), "Edit", map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		if len(values) != 6 {
			t.Errorf("fields length %d is not what expected: %v", len(values), values)
		}
	})

	//missing select, sqlValue, sqlHidden and maybe boolean
	t.Run("Do a search and process results", func(t *testing.T) {
		values, _, _, err := db.SearchMatrix(0, db.ModuleGetIdByName(tableName), "SELECT * FROM "+tableName, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(values) != 4 || values[3]["defaultModuleId"] != "ejaCommands" {
			t.Errorf("Search Matrix problem: %v", values)
		}
	})

	t.Run("Extract a search query", func(t *testing.T) {
		query, args, err := db.SearchQuery(2, tableName, map[string]string{"username": "user"})
		if err != nil {
			t.Fatal(err)
		}
		if len(args) != 1 {
			t.Errorf("Seach query didn't return enough args: %s %v", query, args)
		}
		rows, err := db.Rows(query, args...)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Errorf("Search query result didn't match number of rows: %v", rows)
		}
	})
}
