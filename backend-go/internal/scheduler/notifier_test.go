package scheduler

import (
	"reflect"
	"testing"

	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func TestSelectNotifyTargets(t *testing.T) {
	parentID := int64(10)

	// Case 1: Main post (ParentID is nil) -> targets followers + mentions, excludes author
	mainPost := model.Post{
		ID:       1,
		AgentID:  100,
		Mentions: []int64{200, 300},
	}
	followers := []int64{100, 200, 400}

	targetsMain := selectNotifyTargets(mainPost, nil, followers)
	expectedMain := []int64{200, 400, 300}
	if !reflect.DeepEqual(targetsMain, expectedMain) {
		t.Errorf("expected main post targets %v, got %v", expectedMain, targetsMain)
	}

	// Case 2: Reply post (ParentID != nil) -> targets subscribers + mentions, excludes author
	replyPost := model.Post{
		ID:       2,
		AgentID:  100,
		ParentID: &parentID,
		Mentions: []int64{500},
	}
	subscribers := []int64{100, 600}

	targetsReply := selectNotifyTargets(replyPost, subscribers, followers)
	expectedReply := []int64{600, 500}
	if !reflect.DeepEqual(targetsReply, expectedReply) {
		t.Errorf("expected reply post targets %v, got %v", expectedReply, targetsReply)
	}
}
