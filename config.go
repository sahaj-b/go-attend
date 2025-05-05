package main

// TODO: validate and fix according to config
func validateAndFixRecords(records *Records) error { return nil }

func getHeaderFromCfg() (Record, error) {
	// TODO: add cahching
	return Record{"Date", "English", "Math", "Science", "History", "Geography"}, nil
}

func getNewItemsFromCfg(weekday string) ([]Item, error) {
	return []Item{
		{name: "English", selected: false, status: absent},
		{name: "Math", selected: false, status: present},
		{name: "Science", selected: false, status: absent},
		{name: "History", selected: false, status: cancelled},
		{name: "Geography", selected: false, status: present},
	}, nil
}
