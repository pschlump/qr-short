package storage

// Copyright (C) Philip Schlump 2018-2019.

// PersistentData is the specification for the storage
// backend for the qr-shortner.  Two different interfaces
// are supplied, a file system interface and an interface
// to redis.
type PersistentData interface {
	Insert(URL string) (ID string, err error)
	Exists(ID string) (found bool)
	Update(ULR string, ID string) (codeID string, err error)
	Fetch(ID string) (URL string, err error)
	FetchRaw(ID string) (URL string, err error)
	NextID() (ID string)
	List(string, string) ([]ListData, error)
	UpdateInsert(URL string, ID string) (ur UpdateRespItem)
	IncrementRedirectCount(id string)
}

// ListData is used to format the data returned by the /list API
// end point into a JSON data.
type ListData struct {
	ID    string `json:"Id"`
	URL   string `json:"URL"`
	Count int    `json:"Count"`
}

// UpdateRespItem is a output type used to respond to bulk udpate
// request.  It itimises success/fail and error messages for each
// of the bulk items.
type UpdateRespItem struct {
	ID  string `json:"Id"`
	Msg string `json:"msg"`
	Pos int
}
