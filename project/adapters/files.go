package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)


type FileServiceClient struct {
	clients *clients.Clients
}

func NewFileServiceClient(clients *clients.Clients) *FileServiceClient {
	if clients == nil {
		panic("NewFileServiceClient: clients is nil")
	}
	return &FileServiceClient{
		clients: clients,
	}
}

func (f *FileServiceClient) PutFile(ctx context.Context, fileId string, fileContent string) error {
	res, err := f.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileId, fileContent)
	if err != nil {
		return err
	}

	if res.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).With("file", fileId).Info("file already exists")
		return nil
	}

	if res.StatusCode() != http.StatusCreated {
		return fmt.Errorf("unexpected error code while uploading file %s : %d", fileId, res.StatusCode())
	}

	return nil
}
