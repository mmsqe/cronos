package memiavl

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/ledgerwatch/erigon-lib/mmap"
)

// MmapFile manage the resources of a mmap-ed file
type MmapFile struct {
	file *os.File
	data []byte
	// mmap handle for windows (this is used to close mmap)
	handle *[mmap.MaxMapSize]byte
}

// Open openes the file and create the mmap.
// the mmap is created with flags: PROT_READ, MAP_SHARED, MADV_RANDOM.
func NewMmap(path string) (*MmapFile, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	data, handle, err := Mmap(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	return &MmapFile{
		file:   file,
		data:   data,
		handle: handle,
	}, nil
}

// Close closes the file and mmap handles
func (m *MmapFile) Close() error {
	var err error

	if merr := m.file.Close(); merr != nil {
		err = multierror.Append(err, merr)
	}

	if merr := mmap.Munmap(m.data, m.handle); merr != nil {
		err = multierror.Append(err, merr)
	}

	return err
}

// Data returns the mmap-ed buffer
func (m *MmapFile) Data() []byte {
	return m.data
}
