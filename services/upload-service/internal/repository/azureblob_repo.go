package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type AzureBlobRepositoryImpl struct {
	client        *azblob.Client
	containerName string
}

// NewAzureBlobRepository initializes connection to Azure Blob Storage
// connectionString: Azure storage account connection string
// containerName: Name of the blob container (will be created if doesn't exist)
func NewAzureBlobRepository(connectionString, containerName string) (*AzureBlobRepositoryImpl, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	// Create container if it doesn't exist
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.CreateContainer(ctx, containerName, &container.CreateOptions{})
	if err != nil {
		// Container might already exist, which is fine
		log.Printf("[Azure] Container creation warning: %v", err)
	}

	return &AzureBlobRepositoryImpl{
		client:        client,
		containerName: containerName,
	}, nil
}

// UploadBlob uploads a file to Azure Blob Storage and returns the blob URL
func (r *AzureBlobRepositoryImpl) UploadBlob(ctx context.Context, blobName string, fileData io.Reader, contentType string) (string, error) {
	containerClient := r.client.ServiceClient().NewContainerClient(r.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(blobName)
	fileBytes, err := io.ReadAll(fileData)
	if err != nil {
		return "", fmt.Errorf("failed to read file data: %w", err)
	}

	// Upload blob - Azure SDK requires ReadSeekCloser
	uploadOptions := &blockblob.UploadOptions{}

	// bytes.Reader implements Read and Seek, wrap with NopCloser for Close method
	fileReader := &readSeekCloser{Reader: bytes.NewReader(fileBytes)}

	_, err = blockBlobClient.Upload(ctx, fileReader, uploadOptions)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob %s: %w", blobName, err)
	}

	// Get blob URL
	blobURL := blockBlobClient.URL()
	log.Printf("[Azure] Successfully uploaded blob: %s -> %s", blobName, blobURL)
	return blobURL, nil
}

// DeleteBlob removes a file from Azure Blob Storage
func (r *AzureBlobRepositoryImpl) DeleteBlob(ctx context.Context, blobName string) error {
	containerClient := r.client.ServiceClient().NewContainerClient(r.containerName)
	blockBlobClient := containerClient.NewBlockBlobClient(blobName)

	_, err := blockBlobClient.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob %s: %w", blobName, err)
	}

	log.Printf("[Azure] Successfully deleted blob: %s", blobName)
	return nil
}

// readSeekCloser wraps bytes.Reader to implement io.ReadSeekCloser
type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error {
	return nil
}
