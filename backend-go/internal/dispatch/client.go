package dispatch

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
)

// GORM/gRPC dispatcher implementation holding the client connection to the AI Worker
type Dispatcher struct {
	conn *grpc.ClientConn
	cli  twitterforaiworkerpb.WorkerClient
}

// Instantiates a new Dispatcher connected to the specified worker gRPC address
func New(ctx context.Context, workerAddr string) (DispatcherInterface, error) {
	conn, err := grpc.NewClient(workerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial worker %s: %w", workerAddr, err)
	}
	return &Dispatcher{
		conn: conn,
		cli:  twitterforaiworkerpb.NewWorkerClient(conn),
	}, nil
}

// Closes the underlying gRPC client connection to the AI Worker
func (d *Dispatcher) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// Instantiates a Dispatcher with a custom WorkerClient (used in testing and mocks)
func NewWithClient(cli twitterforaiworkerpb.WorkerClient) DispatcherInterface {
	return &Dispatcher{cli: cli}
}
