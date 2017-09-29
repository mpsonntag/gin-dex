package gindex

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"crypto/sha1"

	"github.com/G-Node/gig"
	log "github.com/Sirupsen/logrus"
)

type IndexBlob struct {
	*gig.Blob
	Repoid       string
	Id           int64
	GinRepoId    int64
	CommitSha    string
	Path         string
	Oid          int64
	IndexingTime time.Time
	Content      string
}

func NewCommitFromGig(gCommit *gig.Commit, repoid string) *IndexCommit {
	commit := &IndexCommit{gCommit, repoid, time.Now()}
	return commit
}

func NewBlobFromGig(gBlob *gig.Blob, repoid string) *IndexBlob {
	// Remember keeping the id
	blob := IndexBlob{Blob: gBlob, Repoid: repoid}
	return &blob
}

type IndexCommit struct {
	*gig.Commit
	GinRepoId    string
	IndexingTime time.Time
}

func BlobFromJson(data []byte) (*IndexBlob, error) {
	bl := &IndexBlob{}
	err := json.Unmarshal(data, bl)
	return bl, err
}

func (c *IndexCommit) ToJson() ([]byte, error) {
	return json.Marshal(c)
}

func (c *IndexCommit) AddToIndex(server *ElServer, index string, id gig.SHA1) error {
	data, err := c.ToJson()
	if err != nil {
		return err
	}
	indexid := sha1.Sum([]byte(c.GinRepoId + id.String()))
	err = AddToIndex(data, server, index, "commit", indexid)
	return err
}

func (bl *IndexBlob) ToJson() ([]byte, error) {
	return json.Marshal(bl)
}

func (bl *IndexBlob) AddToIndex(server *ElServer, index string, id gig.SHA1) error {
	f_type, err := DetermineFileType(bl)
	if err != nil {
		log.Errorf("Could not determine file type: %+v", err)
		return nil
	}
	switch f_type {
	case TEXT:
		log.Debugf("Text file found detected")
		ct, err := ioutil.ReadAll(bl)
		if err != nil {
			log.Errorf("Could not read text file content:%+v", err)
			return err
		}
		bl.Content = string(ct)
	case ODML_XML:
		ct, err := ioutil.ReadAll(bl)
		if err != nil {
			return err
		}
		bl.Content = string(ct)
	}
	data, err := bl.ToJson()
	if err != nil {
		return err
	}
	err = AddToIndex(data, server, index, "blob", id)
	return err
}

func (bl *IndexBlob) IsInIndex() (bool, error) {
	return false, nil
}

func AddToIndex(data []byte, server *ElServer, index, doctype string, id gig.SHA1) error {
	resp, err := server.Index(index, doctype, data, id)
	bd, err := ioutil.ReadAll(resp.Body)
	log.Debugf("Tried adding to the index: %+v, %s", resp, bd)
	return err
}
