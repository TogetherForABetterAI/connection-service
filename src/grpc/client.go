package client

import (
	"google.golang.org/grpc"
	pb "auth-gateway/src/pb/new-client-service"
	"context"
	"google.golang.org/grpc/credentials/insecure"
)


type NotificationClient struct {
	conn   *grpc.ClientConn
	client pb.ClientNotificationServiceClient
}

func NewClient(address string) (*NotificationClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewClientNotificationServiceClient(conn)
	return &NotificationClient{conn: conn, client: client}, nil
}

func (c *NotificationClient) NotifyNewClient(ctx context.Context, newClientRequest *pb.NewClientRequest) error {
	_, err := c.client.NotifyNewClient(ctx, newClientRequest)
	return err
}