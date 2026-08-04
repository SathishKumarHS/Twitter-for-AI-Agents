package sched

import (
	"reflect"
	"testing"

	"github.com/agent-feed/backend-go/internal/db/posts"
)

func intPtr(v int64) *int64 { return &v }

func TestSelectNotifyTargetsMainPost(t *testing.T) {
	created := posts.Post{ID: 10, AgentID: 1, Kind: "post", Mentions: []int64{2, 5}}
	followers := []int64{2, 3, 4}
	got := selectNotifyTargets(created, nil, followers)

	want := []int64{2, 3, 4, 5} // followers + mention 5; mention 2 is a follower (dedup); author 1 excluded
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSelectNotifyTargetsReplyNotifiesSubscribersOnly(t *testing.T) {
	created := posts.Post{ID: 20, AgentID: 3, Kind: "reply", ParentID: intPtr(10), RootID: intPtr(10), Mentions: []int64{7}}
	subscribers := []int64{1, 3, 4} // 3 is the commenter -> excluded
	got := selectNotifyTargets(created, subscribers, []int64{99, 100})

	want := []int64{1, 4, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSelectNotifyTargetsExcludesCommenterAndDedupes(t *testing.T) {
	created := posts.Post{ID: 30, AgentID: 5, Kind: "reply", ParentID: intPtr(10), RootID: intPtr(10), Mentions: []int64{5, 1, 1}}
	got := selectNotifyTargets(created, []int64{1, 5, 2}, nil)

	want := []int64{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSelectNotifyTargetsEmptyAudience(t *testing.T) {
	created := posts.Post{ID: 40, AgentID: 2, Kind: "post"}
	got := selectNotifyTargets(created, nil, nil)
	if len(got) != 0 {
		t.Fatalf("expected empty audience, got %v", got)
	}
}

func TestCleanText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"   hello world   ", "hello world"},
		{"...lead with context", "lead with context"},
		{"…ellipsis lead", "ellipsis lead"},
		{"---dash lead", "dash lead"},
		{string(make([]rune, 400)), string(make([]rune, 280))},
	}
	for _, c := range cases {
		if got := cleanText(c.in); got != c.want {
			t.Errorf("cleanText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
