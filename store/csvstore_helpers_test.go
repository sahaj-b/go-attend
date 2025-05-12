package store

import (
	"reflect"
	"testing"

	"github.com/sahaj-b/go-attend/core"
	"github.com/sahaj-b/go-attend/state"
)

func TestValidateHeader(t *testing.T) {
	tests := []struct {
		header []string
		isErr  bool
	}{
		{[]string{"Date", "English", "Math"}, false},
		{[]string{"Date"}, true},
		{[]string{"Math", "English"}, true},
		{[]string{}, true},
	}

	for _, test := range tests {
		err := validateHeader(test.header)
		if test.isErr && err == nil {
			t.Errorf("Expected error for header %v, got nil", test.header)
		} else if !test.isErr && err != nil {
			t.Errorf("Unexpected error for header %v: %v", test.header, err)
		}
	}
}

func TestValidateRecord(t *testing.T) {
	tests := []struct {
		header []string
		record csvRecord
		isErr  bool
	}{
		{[]string{"Date", "English", "Math"}, csvRecord{"01-10-2023", "1", "0"}, false},
		{[]string{"Date", "English"}, csvRecord{"01/10-2023", "1"}, true},
		{[]string{"Date", "English"}, csvRecord{"01-10-2023", "1", "0"}, true},
		{[]string{"Date", "English"}, csvRecord{}, true},
		{[]string{"Date", "English", "Math"}, csvRecord{"01-10-2023", "3", "0", "2"}, true},
	}

	for _, test := range tests {
		err := validateRecord(test.header, test.record)
		if test.isErr && err == nil {
			t.Errorf("Expected error for record %v, got nil", test.record)
		} else if !test.isErr && err != nil {
			t.Errorf("Unexpected error for record %v: %v", test.record, err)
		}
	}
}

func TestItemsToRecord(t *testing.T) {
	headers := []string{"Date", "English", "Math", "Science", "History", "Geography"}
	tests := []struct {
		Name   string
		date   string
		items  []state.Item
		record csvRecord
		isErr  bool
	}{
		{
			Name: "all subjects",
			date: "10-01-2023",
			items: []state.Item{
				{Name: "English", Selected: false, Status: core.Present},
				{Name: "Math", Selected: false, Status: core.Absent},
				{Name: "Science", Selected: false, Status: core.Present},
				{Name: "History", Selected: false, Status: core.Absent},
				{Name: "Geography", Selected: false, Status: core.Present},
			},
			record: csvRecord{"10-01-2023", "0", "1", "0", "1", "0"},
			isErr:  false,
		},
		{
			Name: "missing subjects",
			date: "10-02-2023",
			items: []state.Item{
				{Name: "Math", Selected: false, Status: core.Present},
				{Name: "History", Selected: false, Status: core.Absent},
				{Name: "English", Selected: false, Status: core.Present},
			},
			record: csvRecord{"10-02-2023", "0", "0", "", "1", ""},
			isErr:  false,
		},
		{
			Name: "unordered subjects",
			date: "01-03-2025",
			items: []state.Item{
				{Name: "Math", Selected: false, Status: core.Absent},
				{Name: "Geography", Selected: false, Status: core.Present},
				{Name: "Science", Selected: false, Status: core.Present},
				{Name: "English", Selected: false, Status: core.Present},
				{Name: "History", Selected: false, Status: core.Absent},
			},
			record: csvRecord{"01-03-2025", "0", "1", "0", "1", "0"},
			isErr:  false,
		},
		{
			Name: "invalid Status",
			date: "01-04-2025",
			items: []state.Item{
				{Name: "English", Selected: false, Status: core.AttendanceStatus(10)},
				{Name: "Math", Selected: false, Status: core.Absent},
				{Name: "Science", Selected: false, Status: core.Present},
				{Name: "History", Selected: false, Status: core.Absent},
				{Name: "Geography", Selected: false, Status: core.Present},
			},
			record: nil,
			isErr:  true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			record, err := stateItemsToRecordStr(headers, test.date, test.items)
			if test.isErr && err == nil {
				t.Errorf("Expected error for %s, got nil", test.Name)
			} else if !test.isErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", test.Name, err)
			} else if !test.isErr && record != nil && !reflect.DeepEqual(record, test.record) {
				t.Errorf("%s: Expected %v, got %v", test.Name, test.record, record)
			}
		})
	}
}

func TestRecordToStateItems(t *testing.T) {
	tests := []struct {
		Name    string
		headers []string
		record  csvRecord
		items   []state.Item
		isErr   bool
	}{
		{
			Name:    "valid record",
			headers: []string{"Date", "Math", "English", "History"},
			record:  csvRecord{"10-01-2023", "0", "1", "2"},
			items: []state.Item{
				{Name: "Math", Selected: false, Status: core.Present},
				{Name: "English", Selected: false, Status: core.Absent},
				{Name: "History", Selected: false, Status: core.Cancelled},
			},
			isErr: false,
		},
		{
			Name:    "all attended",
			headers: []string{"Date", "Math", "English", "History"},
			record:  csvRecord{"10-02-2023", "0", "0", "0"},
			items: []state.Item{
				{Name: "Math", Selected: false, Status: core.Present},
				{Name: "English", Selected: false, Status: core.Present},
				{Name: "History", Selected: false, Status: core.Present},
			},
			isErr: false,
		},
		{
			Name:    "missing subject",
			headers: []string{"Date", "Math", "English", "History"},
			record:  csvRecord{"10-04-2023", "0", "1"},
			items: []state.Item{
				{Name: "Math", Selected: false, Status: core.Present},
				{Name: "English", Selected: false, Status: core.Absent},
			},
			isErr: false,
		},
		{
			Name:    "extra field",
			headers: []string{"Date", "Math", "English"},
			record:  csvRecord{"10-05-2023", "0", "1", "0", "1"},
			items:   nil,
			isErr:   true,
		},
		{
			Name:    "invalid Status value",
			headers: []string{"Date", "Math", "English", "History"},
			record:  csvRecord{"10-06-2023", "10", "1", "0"},
			items:   nil,
			isErr:   true,
		},
		{
			Name:    "empty record",
			headers: []string{"Date", "Math", "English", "History"},
			record:  csvRecord{},
			items:   nil,
			isErr:   true,
		},
		{
			Name:    "empty headers",
			headers: []string{},
			record:  csvRecord{"10-07-2023", "0", "1", "0"},
			items:   nil,
			isErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			items, err := recordStrToStateItems(test.headers, test.record)
			if test.isErr && err == nil {
				t.Errorf("%v: Expected error for record %v, got nil", test.Name, test.record)
			} else if !test.isErr && err != nil {
				t.Errorf("%v: Unexpected error for record %v: %v", test.Name, test.record, err)
			} else if !test.isErr && items != nil && !reflect.DeepEqual(items, test.items) {
				t.Errorf("%v: Expected %v, got %v", test.Name, test.items, items)
			}
		})
	}
}

// TODO: TestValidate
// func TestValidate(t *testing.T) {
// }
