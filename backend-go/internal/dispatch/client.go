package dispatch

import (
	"context"
	"fmt"

	agentfeedpb "github.com/agent-feed/backend-go/gen/agentfeedpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Dispatcher struct {
	conn *grpc.ClientConn
	cli  agentfeedpb.WorkerClient
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
		cli:  agentfeedpb.NewWorkerClient(conn),
	}, nil
}


func (d *Dispatcher) Close() error {
	return d.conn.Close()
}