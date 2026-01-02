package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// AzureBlobRepositoryImpl implements AzureBlobRepository for read-service
type AzureBlobRepositoryImpl struct {
	client        *azblob.Client
	containerName string
}

// NewAzureBlobRepository creates a new AzureBlobRepositoryImpl
func NewAzureBlobRepository(connectionString, containerName string) (*AzureBlobRepositoryImpl, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure blob client: %w", err)
	}

	return &AzureBlobRepositoryImpl{
		client:        client,
		containerName: containerName,
	}, nil
}

// GetBlob downloads the blob with the given name and returns its bytes
func (r *AzureBlobRepositoryImpl) GetBlob(ctx context.Context, blobName string) ([]byte, error) {
	// create a context with timeout to avoid hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	containerClient := r.client.ServiceClient().NewContainerClient(r.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(blobName)

	resp, err := blockBlobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob %s: %w", blobName, err)
	}

	// resp.Body may be an io.ReadCloser or provide a Body method; try to read directly
	var reader io.ReadCloser

	// Several SDK versions expose Body directly.
	if resp.Body != nil {
		reader = resp.Body
	}

	if reader == nil {
		// Fallback: try to call NewRetryReader if available (older SDKs)
		if rr := resp.NewRetryReader; rr != nil {
			reader = rr(ctx, nil)
		}
	}

	if reader == nil {
		return nil, fmt.Errorf("no readable body available for blob %s", blobName)
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("[Azure] error reading blob body: %v", err)
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}

	return data, nil
}
