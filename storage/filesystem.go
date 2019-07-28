package storage

// Copyright (C) Philip Schlump 2018-2019.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/pschlump/godebug"
)

// FileStorage implements PersistentData using the local file system.
type FileStorage struct {
	StorageDir string
	Log        *os.File
	lock       sync.RWMutex
}

// NewFilesystem creates a new connection to the filesystem for storing shorened URLs
func NewFilesystem(storageDir string, log *os.File) (rv PersistentData, err error) {
	err = os.MkdirAll(storageDir, 0744)
	return &FileStorage{
		StorageDir: storageDir,
		Log:        log,
	}, err
}

// NextID returns the next higher integer that will be used to lookup the URL.
func (fs *FileStorage) NextID() string {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	files, err := ioutil.ReadDir(fs.StorageDir)
	if err != nil {
		fmt.Fprintf(fs.Log, "Error: %s, %s\n", err, godebug.LF())
		return ""
	}
	return strconv.FormatUint(uint64(len(files)+1), 36) // Base 36, take count of # of files add 1, this is the ID.
}

// Exists returns true if the ID exists in the file store.
func (fs *FileStorage) Exists(ID string) bool {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fn := filepath.Join(fs.StorageDir, ID)
	if FileExists(fn) {
		return true
	}
	return false
}

// Insert writes out the `urlStr` into the `~/data` direcotry under the file name in `ID`
func (fs *FileStorage) Insert(urlStr string) (string, error) {
	id := fs.NextID()
	return fs.Update(urlStr, id)
}

// Update update an existing URL encode.
func (fs *FileStorage) Update(urlStr, id string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	err := ioutil.WriteFile(filepath.Join(fs.StorageDir, id), []byte(urlStr), 0644)
	if err != nil {
		fmt.Fprintf(fs.Log, "Error: writing file: %s\n", err)
		return id, err
	}
	return id, nil
}

// Fetch converts from a `id` into a `url` to be returned.
func (fs *FileStorage) Fetch(id string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	data, err := ioutil.ReadFile(filepath.Join(fs.StorageDir, id))
	return string(data), err
}

func (fs *FileStorage) FetchRaw(id string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	data, err := ioutil.ReadFile(filepath.Join(fs.StorageDir, id))
	return string(data), err
}

// List returns a list of all the redirects between beg and end, where beg
// can be 0 to start at the beginning and end can be 'last' to to go the most
// recent item.
func (fs *FileStorage) List(beg, end string) (dat []ListData, err error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	// data, err := ioutil.ReadFile(filepath.Join(fs.StorageDir, id))
	// xyzzy - direcotyr list, pull out stuff
	return
}

// UpdateInsert performs an Insert if the ID is new or an Update if it already exists.
func (fs *FileStorage) UpdateInsert(URL string, ID string) (ur UpdateRespItem) {

	var err error
	var code string

	ur.ID = code
	fn := filepath.Join(fs.StorageDir, ID)
	if !FileExists(fn) {
		code, err = fs.Update(URL, ID)
		ur.Msg = "success/insert"
	} else {
		code, err = fs.Update(URL, ID)
		ur.Msg = "success/update"
	}

	if err != nil {
		ur.Msg = fmt.Sprintf("fail:%s", err)
	}

	return
}

// FileExists returns true if the file exists in the file system.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (fs *FileStorage) IncrementRedirectCount(id string) {
	// xyzzy - need to implement this.
}
