package block

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/uuid"
)

type WAL interface {
	AllBlocks() ([]HeadBlock, error)
	NewBlock(id uuid.UUID, instanceID string) (HeadBlock, error)
}

type wal struct {
	c Config
}

func New(c Config) (WAL, error) {
	// make folder
	err := os.MkdirAll(c.Filepath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &wal{
		c: c,
	}, nil
}

func (w *wal) AllBlocks() ([]HeadBlock, error) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s", w.c.Filepath))
	if err != nil {
		return nil, err
	}

	blocks := make([]HeadBlock, 0, len(files))
	for _, f := range files {
		name := f.Name()
		blockID, instanceID, err := parseFilename(name)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, &headBlock{
			completeBlock: completeBlock{
				filepath:   w.c.Filepath,
				blockID:    blockID,
				instanceID: instanceID,
			},
		})
	}

	return blocks, nil
}

func (w *wal) NewBlock(id uuid.UUID, instanceID string) (HeadBlock, error) {
	name := fullFilename(w.c.Filepath, id, instanceID)

	_, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return &headBlock{
		completeBlock: completeBlock{
			filepath:   w.c.Filepath,
			blockID:    id,
			instanceID: instanceID,
		},
	}, nil
}

func parseFilename(name string) (uuid.UUID, string, error) {
	i := strings.Index(name, ":")

	if i < 0 {
		return uuid.UUID{}, "", fmt.Errorf("unable to parse %s", name)
	}

	blockIDString := name[:i]
	instanceID := name[i+1:]

	blockID, err := uuid.Parse(blockIDString)
	if err != nil {
		return uuid.UUID{}, "", err
	}

	return blockID, instanceID, nil
}

func fullFilename(filepath string, blockID uuid.UUID, instanceID string) string {
	return fmt.Sprintf("%s/%v:%v", filepath, blockID, instanceID)
}