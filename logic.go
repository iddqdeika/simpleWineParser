package main

import (
	"errors"
)

//structured object for simplier interpreting data for writing excel sheet.
//has list of columns with their nums
type Table struct {
	Columns map[string]int
	Rows    []Row
}

//substruct of table, contains cells. key of map is num if column
type Row struct {
	Cells map[int]string
}

//set cell value for current row.
func (t *Table) SetCellValue(column string, value string) {
	if !t.ContainsColumn(column) {
		t.AddColumn(column)
	}
	t.Rows[len(t.Rows)-1].Cells[t.Columns[column]] = value
}

//check has table given column yet or not
func (t *Table) ContainsColumn(column string) bool {
	_, ok := t.Columns[column]
	if ok {
		return true
	}
	return false
}

//add column to table
func (t *Table) AddColumn(column string) {
	if t.Columns == nil {
		t.Columns = make(map[string]int)
	}
	t.Columns[column] = len(t.Columns)
}

//add new row to table
func (t *Table) AddRow() error {
	if len(t.Rows) < 1048576 {
		row := Row{}
		row.Cells = make(map[int]string)
		t.Rows = append(t.Rows, row)
		return nil
	}
	return errors.New("Not wnough rows count.")
}
