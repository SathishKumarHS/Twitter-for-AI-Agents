package dispatch

import (
	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

// Helper: converts a Go db model.Agent into a Protobuf Agent message
func toProtoAgent(a model.Agent) *twitterforaiworkerpb.Agent {
	return &twitterforaiworkerpb.Agent{
		Id:        a.ID,
		Handle:    a.Handle,
		Name:      a.Name,
		Persona:   a.Persona,
		Goals:     a.Goals,
		Policies:  a.Policies,
		Interests: a.Interests,
		AvatarUrl: a.AvatarURL,
		Active:    a.Active,
	}
}

// Helper: converts a Go db model.Post into a Protobuf Post message
func toProtoPost(p model.Post) *twitterforaiworkerpb.Post {
	parent := int64(0)
	if p.ParentID != nil {
		parent = *p.ParentID
	}
	root := int64(0)
	if p.RootID != nil {
		root = *p.RootID
	}
	return &twitterforaiworkerpb.Post{
		Id:           p.ID,
		AgentId:      p.AgentID,
		Body:         p.Body,
		Kind:         p.Kind,
		ParentId:     parent,
		RootId:       root,
		Mentions:     p.Mentions,
		AuthorHandle: p.AuthorHandle,
		AuthorName:   p.AuthorName,
	}
}
