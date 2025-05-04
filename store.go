package main

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type Record []string
type Records [][]string
type RecordMap map[string]string

var DATE_FORMAT_STORE = "02-01-2006"

// 0: absent, 1: attended, 2: cancelled
var itemStatusArr = [3]ItemStatus{absent, attended, cancelled}

func (item ItemStatus) statusNum() string {
	for i, status := range itemStatusArr {
		if status == item {
			return strconv.Itoa(i)
		}
	}
	return "-1"
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

func validateRecord(header []string, record Record) error {
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
		if statusNum, err := strconv.Atoi(val); err != nil || statusNum >= len(itemStatusArr) {
			return errors.New("Invalid status number: " + val)
		}
	}
	return nil
}

func itemsToRecord(header Record, date string, items []Item) (Record, error) {
	validateHeader(header)
	if len(items) != len(header)-1 {
		return nil, errors.New("Items length must match header length - 1")
	}
	record := make(Record, len(items)+1)
	record[0] = date
	for _, item := range items {
		statusNum := item.status.statusNum()
		if statusNum == "-1" {
			return nil, errors.New("Invalid item status: " + item.status.text)
		}
		match := false
		for j, subject := range header[1:] {
			if item.name == subject {
				record[j+1] = statusNum
				match = true
				break
			}
		}
		if !match {
			return nil, errors.New("No '" + item.name + "' in header")
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

func getDataFile(mode string) (dataFile *os.File, err error) {
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

func getAllRecords() (Records, error) {
	file, err := getDataFile("r")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return Records(records), nil
}

// TODO: validate according to config
func (record *Records) validate() error { return nil }

func recordToItems(header []string, record Record) ([]Item, error) {
	if err := validateHeader(header); err != nil {
		return nil, errors.New("Invalid header: " + err.Error())
	}
	if err := validateRecord(header, record); err != nil {
		return nil, errors.New("Invalid record: " + err.Error())
	}
	items := make([]Item, len(header)-1)
	for i, header := range header[1:] {
		statusNum, _ := strconv.Atoi(record[i+1]) // validateRecord() ensures this is safe
		items[i] = Item{
			name:     header,
			selected: false,
			status:   itemStatusArr[statusNum],
		}
	}
	return items, nil
}

func (records *Records) getItemsByDate(date string) ([]Item, bool) {
	for _, record := range *records {
		if record[0] == date {
			items, err := recordToItems((*records)[0], record)
			if err != nil {
				panic("Failed to get record: " + err.Error())
			}
			return items, true
		}
	}
	return nil, false
}

func ensureHeader(file *os.File) error {
	if _, err := file.Seek(0, 0); err != nil {
		return errors.New("Failed to seek: " + err.Error())
	}
	reader := csv.NewReader(file)
	_, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			file.Seek(0, 0)
			writer := csv.NewWriter(file)
			defer writer.Flush()
			header, err := getHeaderFromCfg()
			if err != nil {
				return err
			}
			if err := writer.Write(header); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func addItems(date string, items []Item) error {
	header, err := getHeaderFromCfg()
	if err != nil {
		return errors.New("Failed to get header from config: " + err.Error())
	}
	record, err := itemsToRecord(header, date, items)
	if err != nil {
		return err
	}

	file, err := getDataFile("rw")
	if err != nil {
		return err
	}
	defer file.Close()
	if err := ensureHeader(file); err != nil {
		return errors.New("Failed to ensure header: " + err.Error())
	}

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return errors.New("Failed to seek: " + err.Error())
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write(record); err != nil {
		return err
	}
	return nil
}

func writeRecords(records *Records) error {
	file, err := getDataFile("w")
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

func (records *Records) updateItems(date string, items []Item) error {
	for i, record := range *records {
		if record[0] == date {
			header, err := getHeaderFromCfg()
			if err != nil {
				return errors.New("Failed to get header from config: " + err.Error())
			}
			newRecord, err := itemsToRecord(header, date, items)
			if err != nil {
				return err
			}
			(*records)[i] = newRecord
		}
		return writeRecords(records)
	}
	return errors.New("Record not found for: " + date)
}

func (records *Records) handleSave(date string, items []Item) error {
	if _, found := records.getItemsByDate(date); found {
		return records.updateItems(date, items)
	} else {
		return addItems(date, items)
	}
}
