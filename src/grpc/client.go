package client

import (
	"google.golang.org/grpc"
	pb "auth-gateway/src/pb/new-client-service"
	"context"
)


type Client struct {
	conn   *grpc.ClientConn
	client pb.ClientNotificationServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewClientNotificationServiceClient(conn)
	return &Client{conn: conn, client: client}, nil
}

func (c *Client) NotifyNewClient(ctx context.Context, newClientRequest *pb.NewClientRequest) error {
	_, err := c.client.NotifyNewClient(ctx, newClientRequest)
	return err
}