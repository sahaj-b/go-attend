package store

import (
	"encoding/csv"
	"errors"
	"github.com/sahaj-b/go-attend/config"
	"github.com/sahaj-b/go-attend/state"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type CSVStore struct {
	filePath      string
	cachedRecords config.Records
	cacheValid    bool
}

var DATE_FORMAT_STORE = "02-01-2006"

func NewCSVStore() (*CSVStore, error) {
	filePath, err := getDataFilePath()
	if err != nil {
		return nil, errors.New("Failed to get data file path: " + err.Error())
	}
	return &CSVStore{
		filePath:      filePath,
		cachedRecords: make(config.Records, 0),
		cacheValid:    false,
	}, nil
}

func numStrToStatus(s string) (state.ItemStatus, error) {
	num, err := strconv.Atoi(s)
	if err != nil {
		return state.ItemStatus{}, errors.New("invalid status number string: " + s)
	}
	switch state.StatusKind(num) {
	case state.PresentStatus:
		return state.Present, nil
	case state.AbsentStatus:
		return state.Absent, nil
	case state.CancelledStatus:
		return state.Cancelled, nil
	default:
		return state.ItemStatus{}, errors.New("invalid status number: " + s)
	}
}

func validateHeader(header []string) error {
	if len(header) < 2 {
		return errors.New("Header length must be at least 2")
	}
	if header[0] != "Date" {
		return errors.New("First header must be 'Date'")
	}
	return nil
}

func validateRecord(header []string, record config.Record) error {
	// assumes header is valid
	if len(record) < 2 {
		return errors.New("Record length must be at least 2")
	}
	if len(record) != len(header) {
		return errors.New("Record length must match header length")
	}
	if _, err := time.Parse(DATE_FORMAT_STORE, record[0]); err != nil {
		return errors.New("Invalid date format: " + record[0])
	}
	for _, val := range record[1:] {
		// TODO: fix hardcoded max status
		if statusNum, err := strconv.Atoi(val); err != nil || statusNum >= 2 {
			return errors.New("Invalid status number: " + val)
		}
	}
	return nil
}

func itemsToRecordStr(header config.Record, dateStr string, items []state.Item) (config.Record, error) {
	err := validateHeader(header)
	if err != nil {
		return nil, errors.New("Invalid header: " + err.Error())
	}
	if len(items) != len(header)-1 {
		return nil, errors.New("Items length must match header length - 1")
	}
	record := make(config.Record, len(items)+1)
	record[0] = dateStr
	for _, item := range items {
		statusNum := strconv.Itoa(int(item.Status.Kind))
		if statusNum == "-1" {
			return nil, errors.New("Invalid item Kind: " + statusNum)
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
			return nil, errors.New("No '" + item.Name + "' in header")
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
		return "", errors.New("Unsupported OS: " + currOS)
	}
	return path, nil
}

func (cs *CSVStore) getDataFile(mode string) (dataFile *os.File, err error) {
	storePath, err := getDataFilePath()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Dir(storePath), 0755)
	if err != nil {
		return nil, err
	}
	flags := os.O_CREATE
	switch mode {
	case "r":
		flags |= os.O_RDONLY
	case "w":
		flags |= os.O_WRONLY
	case "a":
		flags |= os.O_APPEND | os.O_WRONLY
	case "rw":
		flags |= os.O_RDWR
	default:
		return nil, errors.New("Invalid mode: " + mode)
	}

	dataFile, err = os.OpenFile(storePath, flags, 0644)
	if err != nil {
		return nil, err
	}
	return dataFile, nil
}

func (cs *CSVStore) GetHeader() (config.Record, error) {
	records, err := cs.GetAllRecords()
	if err != nil {
		return nil, errors.New("Failed to fetch records: " + err.Error())
	}
	return records[0], nil
}

func (cs *CSVStore) GetAllRecords() (config.Records, error) {
	if cs.cacheValid {
		return cs.cachedRecords, nil
	}
	file, err := cs.getDataFile("r")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	csvrecords, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	records := config.Records(csvrecords)
	if err := config.ValidateAndFixRecords(&records); err != nil {
		return nil, errors.New("Corrupted data file: " + err.Error())
	}
	cs.cachedRecords = records
	cs.cacheValid = true
	return records, nil
}

func recordStrToItems(header []string, record config.Record) ([]state.Item, error) {
	if err := validateHeader(header); err != nil {
		return nil, errors.New("Invalid header: " + err.Error())
	}
	if err := validateRecord(header, record); err != nil {
		return nil, errors.New("Invalid record: " + err.Error())
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
		return nil, false, errors.New("Failed to fetch records: " + err.Error())
	}
	dateStr := date.Format(DATE_FORMAT_STORE)
	for _, record := range records {
		if record[0] == dateStr {
			items, err := recordStrToItems((records)[0], record)
			if err != nil {
				return nil, false, errors.New("Couldn't convert record to Items: " + err.Error())
			}
			return items, true, nil
		}
	}
	return nil, false, nil
}

func (cs *CSVStore) writeAllRecords(records *config.Records) error {
	file, err := cs.getDataFile("w")
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
		return errors.New("Failed to fetch records: " + err.Error())
	}
	// TODO: handle header validation
	header, err := config.GetHeader()
	if err != nil {
		return errors.New("Failed to get header from config: " + err.Error())
	}

	recordsMap := make(map[string]config.Record)
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
			return errors.New("Failed to convert items to record: " + err.Error())
		}
		if err := validateRecord(header, newRecord); err != nil {
			return errors.New("Invalid record: " + err.Error())
		}
		recordsMap[formattedDate] = newRecord
	}

	finalRecords := make(config.Records, 0, len(recordsMap)+1)
	finalRecords = append(finalRecords, header)
	for _, record := range recordsMap {
		finalRecords = append(finalRecords, record)
	}
	err = cs.writeAllRecords(&finalRecords)
	if err != nil {
		return errors.New("Failed to write records: " + err.Error())
	}
	cs.cachedRecords = finalRecords
	cs.cacheValid = true

	return nil
}

func (cs *CSVStore) SaveState(s *state.State) error {
	s.CachedDates[s.Date] = s.Items
	err := cs.saveRecords(&s.CachedDates)
	if err != nil {
		return errors.New("Failed to save: " + err.Error())
	}
	return nil
}
