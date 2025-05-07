package main

import (
	"reflect"
	"testing"
)

func TestStatusNum(t *testing.T) {
	tests := []struct {
		item   ItemStatus
		status string
	}{
		{present, "1"},
		{absent, "0"},
		{cancelled, "2"},
	}

	for _, test := range tests {
		if test.item.statusNum() != test.status {
			t.Errorf("Expected %s, got %s", test.status, test.item.statusNum())
		}
	}
}

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
		record Record
		isErr  bool
	}{

		{[]string{"Date", "English", "Math"}, Record{"01-10-2023", "1", "0"}, false},
		{[]string{"Date", "English"}, Record{"01/10-2023", "1"}, true},
		{[]string{"Date", "English"}, Record{"01-10-2023", "1", "0"}, true},
		{[]string{"Date", "English"}, Record{}, true},
		{[]string{"Date", "English", "Math"}, Record{"01-10-2023", "3", "0", "2"}, true},
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
		name   string
		date   string
		items  []Item
		record Record
		isErr  bool
	}{
		{
			name: "all subjects",
			date: "10-01-2023",
			items: []Item{
				{name: "English", selected: false, status: present},
				{name: "Math", selected: false, status: absent},
				{name: "Science", selected: false, status: present},
				{name: "History", selected: false, status: absent},
				{name: "Geography", selected: false, status: present},
			},
			record: Record{"10-01-2023", "1", "0", "1", "0", "1"},
			isErr:  false,
		},
		{
			name: "missing subjects",
			date: "10-02-2023",
			items: []Item{
				{name: "Math", selected: false, status: present},
				{name: "History", selected: false, status: absent},
				{name: "English", selected: false, status: present},
			},
			record: nil,
			isErr:  true,
		},
		{
			name: "unordered subjects (no error)",
			date: "01-03-2025",
			items: []Item{
				{name: "Math", selected: false, status: absent},
				{name: "Geography", selected: false, status: present},
				{name: "Science", selected: false, status: present},
				{name: "English", selected: false, status: present},
				{name: "History", selected: false, status: absent},
			},
			record: Record{"01-03-2025", "1", "0", "1", "0", "1"},
			isErr:  false,
		},
		{
			name: "invalid status",
			date: "01-04-2025",
			items: []Item{
				{name: "English", selected: false, status: ItemStatus{text: "hello"}},
				{name: "Math", selected: false, status: absent},
				{name: "Science", selected: false, status: present},
				{name: "History", selected: false, status: absent},
				{name: "Geography", selected: false, status: present},
			},
			record: nil,
			isErr:  true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			record, err := itemsToRecordStr(headers, test.date, test.items)
			if test.isErr && err == nil {
				t.Errorf("Expected error for %s, got nil", test.name)
			} else if !test.isErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", test.name, err)
			} else if !test.isErr && record != nil && !reflect.DeepEqual(record, test.record) {
				t.Errorf("Expected %v, got %v", test.record, record)
			}
		})
	}
}

func TestRecordToItems(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		record  Record
		items   []Item
		isErr   bool
	}{
		{
			name:    "valid record",
			headers: []string{"Date", "Math", "English", "History"},
			record:  Record{"10-01-2023", "1", "0", "2"},
			items: []Item{
				{name: "Math", selected: false, status: present},
				{name: "English", selected: false, status: absent},
				{name: "History", selected: false, status: cancelled},
			},
			isErr: false,
		},
		{
			name:    "all attended",
			headers: []string{"Date", "Math", "English", "History"},
			record:  Record{"10-02-2023", "1", "1", "1"},
			items: []Item{
				{name: "Math", selected: false, status: present},
				{name: "English", selected: false, status: present},
				{name: "History", selected: false, status: present},
			},
			isErr: false,
		},
		{
			name:    "missing subject",
			headers: []string{"Date", "Math", "English", "History"},
			record:  Record{"10-04-2023", "1", "0"},
			items:   nil,
			isErr:   true,
		},
		{
			name:    "extra field",
			headers: []string{"Date", "Math", "English"},
			record:  Record{"10-05-2023", "1", "0", "1", "0"},
			items:   nil,
			isErr:   true,
		},
		{
			name:    "invalid status value",
			headers: []string{"Date", "Math", "English", "History"},
			record:  Record{"10-06-2023", "10", "0", "1"},
			items:   nil,
			isErr:   true,
		},
		{
			name:    "empty record",
			headers: []string{"Date", "Math", "English", "History"},
			record:  Record{},
			items:   nil,
			isErr:   true,
		},
		{
			name:    "empty headers",
			headers: []string{},
			record:  Record{"10-07-2023", "1", "0", "1"},
			items:   nil,
			isErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items, err := recordStrToItems(test.headers, test.record)
			if test.isErr && err == nil {
				t.Errorf("Expected error for record %v, got nil", test.record)
			} else if !test.isErr && err != nil {
				t.Errorf("Unexpected error for record %v: %v", test.record, err)
			} else if !test.isErr && items != nil && !reflect.DeepEqual(items, test.items) {
				t.Errorf("Expected %v, got %v", test.items, items)
			}
		})
	}
}

// TODO: TestValidate
// func TestValidate(t *testing.T) {
// }
