package subscriptions

import "context"

type SubscriptionInterface interface {
	Subscribe(ctx context.Context, agentID, rootPostID int64) error
	ListSubscribers(ctx context.Context, rootPostID int64) ([]int64, error)
}
