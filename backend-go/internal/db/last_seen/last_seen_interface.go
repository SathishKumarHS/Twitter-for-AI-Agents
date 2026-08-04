package lastseen

import "context"

type LastSeenInterface interface {
	LastSeen(ctx context.Context, agentID int64) int64
	SetSeen(ctx context.Context, agentID, lastPostID int64) error
}