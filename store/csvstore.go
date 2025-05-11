package store

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"time"

	"github.com/sahaj-b/go-attend/config"
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

var DATE_FORMAT_CSV = "02-01-2006"

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

func validateHeader(header []string) error {
	if len(header) < 2 {
		return fmt.Errorf("Header length must be at least 2")
	}
	if header[0] != "Date" {
		return fmt.Errorf("First header must be 'Date'")
	}
	subjects := make(map[string]bool)
	for _, subject := range header[1:] {
		if _, exists := subjects[subject]; exists {
			return fmt.Errorf("Duplicate subject found: %v", subject)
		}
	}
	return nil
}

func validateRecord(header []string, record csvRecord) error {
	// assumes header is valid 🙏
	if len(record) < 2 {
		return fmt.Errorf("Record length must be at least 2")
	}
	if len(record) > len(header) {
		return fmt.Errorf("Record length is greater than header")
	}
	// not checking length because header may be bigger for new subjects
	// if len(record) != len(header) {
	// 	return fmt.Errorf("Record length must match header length")
	// }
	if _, err := time.Parse(DATE_FORMAT_CSV, record[0]); err != nil {
		return fmt.Errorf("Invalid date format: %v", record[0])
	}
	for _, val := range record[1:] {
		// TODO: fix hardcoded max status
		if statusNum, err := strconv.Atoi(val); val != "" && (err != nil || statusNum > 2) {
			return fmt.Errorf("Invalid status number: %v", statusNum)
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
	record := make(csvRecord, len(header)) // populate with empty strings
	record[0] = dateStr
	for _, item := range items {
		kindAsInt := int(item.Status)
		// TODO: fix hardcoded max status kind
		if kindAsInt < 0 || kindAsInt > 2 {
			return nil, fmt.Errorf("invalid item Kind: %d", kindAsInt)
		}

		idx := slices.Index(header[1:], item.Name)
		if idx == -1 {
			return nil, fmt.Errorf("No '%v' in header", item.Name)
		}
		record[idx+1] = strconv.Itoa(kindAsInt) // enum to string
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

func (cs *CSVStore) getHeaderFromCfg() csvRecord {
	header := []string{}
	for subject := range config.GetAllSubjects() {
		header = append(header, subject)
	}
	// slices.Sort(header)
	header = append([]string{"Date"}, header...)
	return header
}

func (cs *CSVStore) getHeaderFromCsv() (csvRecord, error) {
	records, err := cs.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch records: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("No header found")
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
	reader.FieldsPerRecord = -1
	cs.validateAndUpdateHeader(reader)
	file.Seek(0, 0)
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
	items := []state.Item{}
	for i, subject := range header[1:] {
		if i+1 >= len(record) {
			// record is shorter than header. this means header is updated with new subjects for future records
			// so we can ignore the rest of the header
			break
		}
		if record[i+1] == "" {
			// record is marked as empty. this means the subject is removed from the config.
			// so we can ignore this subject
			continue
		}
		status, err := strconv.Atoi(record[i+1])
		if err != nil {
			return nil, fmt.Errorf("Invalid status number: %v", record[i+1])
		}
		items = append(items, state.Item{
			Name:     subject,
			Selected: false,
			Status:   state.ItemStatus(status),
		})
	}

	return items, nil
}

func (cs *CSVStore) GetItemsByDate(date time.Time) ([]state.Item, bool, error) {
	records, err := cs.GetAllRecords()
	if err != nil {
		return nil, false, fmt.Errorf("Failed to fetch records: %w", err)
	}

	dateStr := date.Format(DATE_FORMAT_CSV)
	for _, record := range records {
		if record[0] == dateStr {
			items, err := recordStrToItems(records[0], record)
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

func (cs *CSVStore) validateAndUpdateHeader(reader *csv.Reader) error {
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("Failed to read records: %w", err)
	}
	if len(records) == 0 {
		correctHeader := cs.getHeaderFromCfg()
		err = cs.writeAllRecords(&csvRecords{correctHeader})
		if err != nil {
			return fmt.Errorf("Failed to write header: %w", err)
		}
		return nil
	}
	if len(records[0]) < 1 {
		return fmt.Errorf("CSV Header length must be at least 1")
	}

	// handling schedule change
	newSubjects := []string{}
	cfgSubjects := config.GetAllSubjects()
	for subject := range cfgSubjects {
		if !slices.Contains(records[0], subject) {
			newSubjects = append(newSubjects, subject)
		}
	}
	newSubjectsNum := len(newSubjects)
	cfgSubjectsNum := len(cfgSubjects)
	oldSubjectsNum := len(records[0]) - 1
	if newSubjectsNum > 0 {
		// if newSubjectsNum = cfgSubjectsNum - oldSubjectsNum; newSubjectsNum > 0 {
		// SUBJECTS ADDED ONLY
		records[0] = append(records[0], newSubjects...)
		csvrecords := csvRecords(records)
		err = cs.writeAllRecords(&csvrecords)
		if err != nil {
			return fmt.Errorf("Failed to write header: %w", err)
		}
		// } else {
		// SUBJECTS ADDED AND REMOVED
		// }
	} else if oldSubjectsNum > cfgSubjectsNum {
		// SUBJECTS REMOVED ONLY
	}

	// NO CHANGE
	return nil
}

func (cs *CSVStore) saveRecords(imap *state.ItemsMap) error {
	allRecords, err := cs.GetAllRecords()
	if err != nil {
		return fmt.Errorf("Failed to fetch records: %w", err)
	}
	header := allRecords[0]
	recordsMap := make(map[string]csvRecord)
	for i, record := range allRecords {
		if i == 0 {
			continue
		}
		recordsMap[record[0]] = record
	}

	for date, items := range *imap {
		formattedDate := date.Format(DATE_FORMAT_CSV)
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
		return err
	}
	return nil
}
