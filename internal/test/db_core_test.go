// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package test

import (
	"testing"
	"tibula/internal/db"
)

func TestDbOperations(t *testing.T) {
	t.Run("DbOpen", func(t *testing.T) {
		err := db.Open("sqlite", ":memory:", "", "", "", 0)
		if err != nil {
			t.Error("Cannot open database:", err)
		}
	})

	t.Run("DbTableAdd", func(t *testing.T) {
		if err := db.TableAdd("table_test"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbFieldAddText", func(t *testing.T) {
		if err := db.FieldAdd("table_test", "name", "text"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbFieldAddInteger", func(t *testing.T) {
		if err := db.FieldAdd("table_test", "age", "integer"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbRunInsertNew", func(t *testing.T) {
		runValues, runError := db.Run("INSERT INTO table_test (ejaOwner, ejaLog) VALUES (?,?)", 1, db.Now())
		if runError != nil || runValues.LastId != 1 {
			t.Error(runValues, runError)
		}
	})

	t.Run("DbRunInsert", func(t *testing.T) {
		runValues, runError := db.Run("INSERT INTO table_test VALUES (NULL,1,?,?,?)", db.Now(), "uno", 100)
		if runError != nil || runValues.LastId != 2 {
			t.Error(runValues, runError)
		}
	})

	t.Run("DbValueSelect", func(t *testing.T) {
		value, valueError := db.Value("SELECT name FROM table_test WHERE ejaId = ?", 2)
		if valueError != nil || value != "uno" {
			t.Error(value, valueError)
		}
	})

	t.Run("DbRowSelect", func(t *testing.T) {
		row, rowError := db.Row("SELECT * FROM table_test WHERE ejaId = ?", 2)
		if rowError != nil || len(row) != 5 {
			t.Error(row, rowError)
		}
	})

	t.Run("DbRowsSelect", func(t *testing.T) {
		rows, rowsError := db.Rows("SELECT * FROM table_test WHERE ejaId > 0")
		if rowsError != nil || len(rows) != 2 {
			t.Error(rows, rowsError)
		}
	})

	t.Run("DbTableDel", func(t *testing.T) {
		if err := db.TableDel("table_test"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbClose", func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
}
