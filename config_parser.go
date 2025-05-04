package main

func getHeaderFromCfg() (Record, error) {
	// TODO: add cahching
	return Record{"Date", "English", "Math", "Science", "History", "Geography"}, nil
}
