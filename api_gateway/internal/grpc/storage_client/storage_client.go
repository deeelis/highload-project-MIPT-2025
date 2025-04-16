package storage_client

import (
	"api_gateway/internal/domain/models"
	"context"
	"fmt"
	"log/slog"
	"time"

	//storagepb "github.com/deeelis/storage-protos/gen/go/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StorageClient struct {
	//client  storagepb.StorageServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
	log     *slog.Logger
}

func NewStorageClient(addr string, timeout time.Duration, log *slog.Logger) (*StorageClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage service: %w", err)
	}

	return &StorageClient{
		//client:  storagepb.NewStorageServiceClient(conn),
		conn:    conn,
		timeout: timeout,
		log:     log,
	}, nil
}

func (c *StorageClient) GetContentStatus(ctx context.Context, contentID string) (*models.ContentStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	//resp, err := c.client.GetContentStatus(ctx, &storagepb.GetContentStatusRequest{
	//	ContentId: contentID,
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get content status: %w", err)
	//}

	return &models.ContentStatus{
		//ID:     resp.ContentId,
		//Status: resp.Status,
		Analysis: &models.AnalysisResult{
			//IsApproved:   resp.IsApproved,
			//IsSpam:       resp.IsSpam,
			//HasSensitive: resp.HasSensitive,
			//Sentiment:    float64(resp.Sentiment),
			//Language:     resp.Language,
		},
	}, nil
}

func (c *StorageClient) Close() error {
	return c.conn.Close()
}
