package store

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/sahaj-b/go-attend/state"
	"github.com/sahaj-b/go-attend/utils"
)

type CSVStore struct {
	filePath      string
	cachedRecords csvRecords
	cacheValid    bool
}

type (
	csvRecord  []string
	csvRecords [][]string
)

var DATE_FORMAT_STORE = "02-01-2006"

func NewCSVStore() (*CSVStore, error) {
	filePath, err := getDataFilePath()
	if err != nil {
		return nil, fmt.Errorf("Failed to get data(csv) file path: %w", err)
	}
	return &CSVStore{
		filePath:      filePath,
		cachedRecords: make(csvRecords, 0),
		cacheValid:    false,
	}, nil
}

func numStrToStatus(s string) (state.ItemStatus, error) {
	num, err := strconv.Atoi(s)
	if err != nil {
		return state.ItemStatus{}, fmt.Errorf("invalid status number string: %v", s)
	}
	switch state.StatusKind(num) {
	case state.PresentStatus:
		return state.Present, nil
	case state.AbsentStatus:
		return state.Absent, nil
	case state.CancelledStatus:
		return state.Cancelled, nil
	default:
		return state.ItemStatus{}, fmt.Errorf("invalid status number: %v", s)
	}
}

func validateHeader(header []string) error {
	if len(header) < 2 {
		return fmt.Errorf("Header length must be at least 2")
	}
	if header[0] != "Date" {
		return fmt.Errorf("First header must be 'Date'")
	}
	return nil
}

func validateRecord(header []string, record csvRecord) error {
	// assumes header is valid
	if len(record) < 2 {
		return fmt.Errorf("Record length must be at least 2")
	}
	if len(record) != len(header) {
		return fmt.Errorf("Record length must match header length")
	}
	if _, err := time.Parse(DATE_FORMAT_STORE, record[0]); err != nil {
		return fmt.Errorf("Invalid date format: %v", record[0])
	}
	for _, val := range record[1:] {
		// TODO: fix hardcoded max status
		if statusNum, err := strconv.Atoi(val); err != nil || statusNum >= 2 {
			return fmt.Errorf("Invalid status number: %v", val)
		}
	}
	return nil
}

func validateRecords(records *csvRecords) error {
	return nil
}

func itemsToRecordStr(header csvRecord, dateStr string, items []state.Item) (csvRecord, error) {
	err := validateHeader(header)
	if err != nil {
		return nil, fmt.Errorf("Invalid header: %w", err)
	}
	if len(items) != len(header)-1 {
		return nil, fmt.Errorf("Items length must match header length - 1")
	}
	record := make(csvRecord, len(items)+1)
	record[0] = dateStr
	for _, item := range items {
		statusNum := strconv.Itoa(int(item.Status.Kind))
		if statusNum == "-1" {
			return nil, fmt.Errorf("Invalid item Kind: %v", statusNum)
		}
		match := false
		for j, subject := range header[1:] {
			if item.Name == subject {
				record[j+1] = statusNum
				match = true
				break
			}
		}
		if !match {
			return nil, fmt.Errorf("No '%v' in header", item.Name)
		}
	}
	return record, nil
}

func getDataFilePath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	currOS := runtime.GOOS
	homeDir := currentUser.HomeDir

	path := ""
	switch currOS {
	case "linux":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(homeDir, ".local", "share")
		}
		path = filepath.Join(dataHome, "go-attend", "attendance.csv")
	case "darwin":
		path = filepath.Join(homeDir, "Library", "Application Support", "go-attend", "attendance.csv")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		path = filepath.Join(appData, "go-attend", "attendance.csv")
	default:
		return "", fmt.Errorf("Unsupported OS: %v", currOS)
	}
	return path, nil
}

func (cs *CSVStore) GetHeader() (csvRecord, error) {
	records, err := cs.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch records: %w", err)
	}
	return records[0], nil
}

func (cs *CSVStore) GetAllRecords() (csvRecords, error) {
	if cs.cacheValid {
		return cs.cachedRecords, nil
	}
	file, err := utils.EnsureAndGetFile(cs.filePath, "r")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	csvrecords, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	records := csvRecords(csvrecords)
	if err := validateRecords(&records); err != nil {
		return nil, fmt.Errorf("Corrupted data file: %w", err)
	}
	cs.cachedRecords = records
	cs.cacheValid = true
	return records, nil
}

func recordStrToItems(header []string, record csvRecord) ([]state.Item, error) {
	if err := validateHeader(header); err != nil {
		return nil, fmt.Errorf("Invalid header: %w", err)
	}
	if err := validateRecord(header, record); err != nil {
		return nil, fmt.Errorf("Invalid record: %w", err)
	}
	items := make([]state.Item, len(header)-1)
	for i, header := range header[1:] {
		status, err := numStrToStatus(record[i+1])
		if err != nil {
			return nil, err
		}
		items[i] = state.Item{
			Name:     header,
			Selected: false,
			Status:   status,
		}
	}
	return items, nil
}

func (cs *CSVStore) GetItemsByDate(date time.Time) ([]state.Item, bool, error) {
	records, err := cs.GetAllRecords()
	if err != nil {
		return nil, false, fmt.Errorf("Failed to fetch records: %w", err)
	}
	dateStr := date.Format(DATE_FORMAT_STORE)
	for _, record := range records {
		if record[0] == dateStr {
			items, err := recordStrToItems((records)[0], record)
			if err != nil {
				return nil, false, fmt.Errorf("Couldn't convert record to Items: %w", err)
			}
			return items, true, nil
		}
	}
	return nil, false, nil
}

func (cs *CSVStore) writeAllRecords(records *csvRecords) error {
	file, err := utils.EnsureAndGetFile(cs.filePath, "w")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.WriteAll(*records); err != nil {
		return err
	}
	return nil
}

func (cs *CSVStore) saveRecords(imap *state.ItemsMap) error {
	allRecords, err := cs.GetAllRecords()
	if err != nil {
		return fmt.Errorf("Failed to fetch records: %w", err)
	}
	// TODO: handle header validation from cfg?
	header := allRecords[0]
	recordsMap := make(map[string]csvRecord)
	for i, record := range allRecords {
		if i == 0 {
			continue
		}
		recordsMap[record[0]] = record
	}

	for date, items := range *imap {
		formattedDate := date.Format(DATE_FORMAT_STORE)
		newRecord, err := itemsToRecordStr(header, formattedDate, items)
		if err != nil {
			return fmt.Errorf("Failed to convert items to record: %w", err)
		}
		if err := validateRecord(header, newRecord); err != nil {
			return fmt.Errorf("Invalid record: %w", err)
		}
		recordsMap[formattedDate] = newRecord
	}

	finalRecords := make(csvRecords, 0, len(recordsMap)+1)
	finalRecords = append(finalRecords, header)
	for _, record := range recordsMap {
		finalRecords = append(finalRecords, record)
	}
	err = cs.writeAllRecords(&finalRecords)
	if err != nil {
		return fmt.Errorf("Failed to write records: %w", err)
	}
	cs.cachedRecords = finalRecords
	cs.cacheValid = true

	return nil
}

func (cs *CSVStore) SaveState(s *state.State) error {
	s.CachedDates[s.Date] = s.Items
	err := cs.saveRecords(&s.CachedDates)
	if err != nil {
		return fmt.Errorf("Failed to save: %w", err)
	}
	return nil
}
