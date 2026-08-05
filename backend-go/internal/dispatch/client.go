package dispatch

import (
	"context"
	"fmt"

	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Dispatcher struct {
	conn *grpc.ClientConn
	cli  twitterforaiworkerpb.WorkerClient
}

func New(ctx context.Context, workerAddr string) (DispatcherInterface, error) {
	conn, err := grpc.DialContext(ctx, workerAddr,
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

func (d *Dispatcher) Close() error {
	return d.conn.Close()
}
