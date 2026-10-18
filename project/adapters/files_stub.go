package adapters

import (
	"context"
	"fmt"
	"sync"
)

type FilesApiStub struct {
	lock sync.Mutex
	files map[string]string
}

func (f *FilesApiStub) PutFile(ctx context.Context, fileId string, fileContent string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.files == nil {
		f.files = make(map[string]string)
	}

	f.files[fileId] = fileContent
	return nil
}

func (f *FilesApiStub) DownloadFile(ctx context.Context, fileId string) (string, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.files ==nil {
		f.files = make(map[string]string)
	}

	content, ok := f.files[fileId]

	if !ok {
		return "", fmt.Errorf("file %s not found", fileId)
	}

	return content, nil

}

