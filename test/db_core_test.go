// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/eja/tibula/db"
)

func TestDbOperations(t *testing.T) {
	session := db.Session()

	dbPath := "./test_concurrent.db"
	os.Remove(dbPath)

	defer func() {
		session.Close()
		os.Remove(dbPath)
	}()

	t.Run("DbOpen", func(t *testing.T) {
		err := session.Open("sqlite", dbPath, "", "", "", 0)
		if err != nil {
			t.Error("Cannot open database:", err)
		}
	})

	t.Run("DbTableAdd", func(t *testing.T) {
		if err := session.TableAdd("table_test"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbFieldAddText", func(t *testing.T) {
		if err := session.FieldAdd("table_test", "name", "text"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbFieldAddInteger", func(t *testing.T) {
		if err := session.FieldAdd("table_test", "age", "integer"); err != nil {
			t.Error(err)
		}
	})

	t.Run("DbRunInsertNew", func(t *testing.T) {
		runValues, runError := session.Run("INSERT INTO table_test (ejaOwner, ejaLog) VALUES (?,?)", 1, session.Now())
		if runError != nil || runValues.LastId != 1 {
			t.Error(runValues, runError)
		}
	})

	t.Run("DbRunInsert", func(t *testing.T) {
		runValues, runError := session.Run("INSERT INTO table_test VALUES (NULL,1,?,?,?)", session.Now(), "uno", 100)
		if runError != nil || runValues.LastId != 2 {
			t.Error(runValues, runError)
		}
	})

	t.Run("DbValueSelect", func(t *testing.T) {
		value, valueError := session.Value("SELECT name FROM table_test WHERE ejaId = ?", 2)
		if valueError != nil || value != "uno" {
			t.Error(value, valueError)
		}
	})

	t.Run("DbRowSelect", func(t *testing.T) {
		row, rowError := session.Row("SELECT * FROM table_test WHERE ejaId = ?", 2)
		if rowError != nil || len(row) != 5 {
			t.Error(row, rowError)
		}
	})

	t.Run("DbRowsSelect", func(t *testing.T) {
		rows, rowsError := session.Rows("SELECT * FROM table_test WHERE ejaId > 0")
		if rowsError != nil || len(rows) != 2 {
			t.Error(rows, rowsError)
		}
	})

	t.Run("ConcurrentStress", func(t *testing.T) {
		tableName := "table_concurrent"
		if err := session.TableAdd(tableName); err != nil {
			t.Fatal(err)
		}
		session.FieldAdd(tableName, "worker_id", "integer")
		session.FieldAdd(tableName, "payload", "text")

		t.Run("ParallelWrites", func(t *testing.T) {
			var wg sync.WaitGroup
			workers := 10
			insertsPerWorker := 50

			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < insertsPerWorker; j++ {
						_, err := session.Run(
							fmt.Sprintf("INSERT INTO %s (ejaOwner, ejaLog, worker_id, payload) VALUES (?,?,?,?)", tableName),
							1, session.Now(), id, fmt.Sprintf("data_%d_%d", id, j),
						)
						if err != nil {
							t.Errorf("Worker %d failed insert: %v", id, err)
							return
						}
					}
				}(i)
			}
			wg.Wait()

			countStr, err := session.Value(fmt.Sprintf("SELECT count(*) FROM %s", tableName))
			if err != nil {
				t.Error(err)
			}
			expected := fmt.Sprintf("%d", workers*insertsPerWorker)
			if countStr != expected {
				t.Errorf("Expected %s rows, got %s", expected, countStr)
			}
		})

		t.Run("MixedReadWrite", func(t *testing.T) {
			var wg sync.WaitGroup
			writers := 5
			readers := 10
			ops := 20

			for r := 0; r < readers; r++ {
				wg.Add(1)
				go func(rid int) {
					defer wg.Done()
					for i := 0; i < ops; i++ {
						_, err := session.Rows(fmt.Sprintf("SELECT * FROM %s LIMIT 5", tableName))
						if err != nil {
							t.Errorf("Reader %d failed: %v", rid, err)
						}
						time.Sleep(2 * time.Millisecond)
					}
				}(r)
			}

			for w := 0; w < writers; w++ {
				wg.Add(1)
				go func(wid int) {
					defer wg.Done()
					for i := 0; i < ops; i++ {
						_, err := session.Run(
							fmt.Sprintf("INSERT INTO %s (ejaOwner, ejaLog, worker_id) VALUES (?,?,?)", tableName),
							1, session.Now(), 999,
						)
						if err != nil {
							t.Errorf("Writer %d failed: %v", wid, err)
						}
					}
				}(w)
			}

			wg.Wait()
		})
	})

	t.Run("DbTableDel", func(t *testing.T) {
		if err := session.TableDel("table_test"); err != nil {
			t.Error(err)
		}
		session.TableDel("table_concurrent")
	})

	t.Run("DbClose", func(t *testing.T) {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	})
}
