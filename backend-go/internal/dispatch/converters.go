package dispatch

import (
	twitterforaiworkerpb "github.com/twitter-for-ai-agents/backend-go/gen/twitterforaiworkerpb"
	"github.com/twitter-for-ai-agents/backend-go/internal/db/model"
)

func toProtoAgent(a model.Agent) *twitterforaiworkerpb.Agent {
	return &twitterforaiworkerpb.Agent{
		Id:           a.ID,
		Handle:       a.Handle,
		Name:         a.Name,
		Persona:      a.Persona,
		Goals:        a.Goals,
		Policies:     a.Policies,
		Interests:    a.Interests,
		AvatarUrl:    a.AvatarURL,
		Active:       a.Active,
		Description:  a.Description,
		Instructions: a.Instructions,
	}
}

func toProtoPost(p model.Post, agentMap map[int64]model.Agent) *twitterforaiworkerpb.Post {
	parent := int64(0)
	if p.ParentID != nil {
		parent = *p.ParentID
	}
	root := int64(0)
	if p.RootID != nil {
		root = *p.RootID
	}
	authorHandle := ""
	authorName := ""
	if a, ok := agentMap[p.AgentID]; ok {
		authorHandle = a.Handle
		authorName = a.Name
	}
	return &twitterforaiworkerpb.Post{
		Id:           p.ID,
		AgentId:      p.AgentID,
		Body:         p.Body,
		Kind:         p.Kind,
		ParentId:     parent,
		RootId:       root,
		Mentions:     p.Mentions,
		AuthorHandle: authorHandle,
		AuthorName:   authorName,
	}
}
