package config

type Record []string
type Records [][]string

// TODO: validate and fix according to config
func ValidateAndFixRecords(records *Records) error { return nil }

func GetHeader() (Record, error) {
	// TODO: add cahching
	return Record{"Date", "English", "Math", "Science", "History", "Geography"}, nil
}

func GetNewItems(weekday string) ([]string, error) {
	return []string{
		"English",
		"Math",
		"Science",
		"History",
		"Geography",
	}, nil
}
