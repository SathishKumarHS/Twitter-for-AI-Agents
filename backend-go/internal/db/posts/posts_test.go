package posts

import (
	"testing"
)

func int64ptr(v int64) *int64 { return &v }

func TestBuildThreadTree(t *testing.T) {
	root := int64(10)
	flat := []Post{
		{ID: 10, Kind: "post"},
		{ID: 11, Kind: "reply", ParentID: int64ptr(10), RootID: int64ptr(10)},
		{ID: 12, Kind: "reply", ParentID: int64ptr(11), RootID: int64ptr(10)},
		{ID: 13, Kind: "reply", ParentID: int64ptr(10), RootID: int64ptr(10)},
	}

	tree := buildThreadTree(flat)
	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	r := tree[0]
	if r.ID != root {
		t.Fatalf("expected root id %d, got %d", root, r.ID)
	}
	if len(r.Comments) != 2 {
		t.Fatalf("expected 2 top-level comments, got %d", len(r.Comments))
	}
	if r.Comments[0].ID != 11 || r.Comments[1].ID != 13 {
		t.Fatalf("unexpected comment order: %d, %d", r.Comments[0].ID, r.Comments[1].ID)
	}
	if len(r.Comments[0].Comments) != 1 || r.Comments[0].Comments[0].ID != 12 {
		t.Fatalf("expected nested comment 12 under 11, got %+v", r.Comments[0].Comments)
	}
}

func TestBuildThreadTreeNoComments(t *testing.T) {
	tree := buildThreadTree([]Post{{ID: 7, Kind: "post"}})
	if len(tree) != 1 || tree[0].ID != 7 || len(tree[0].Comments) != 0 {
		t.Fatalf("unexpected tree: %+v", tree)
	}
}
