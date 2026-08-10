package tools

type allowedTool string

func (a allowedTool) String() string {
	return string(a)
}

const (
	TOOL_GET_FEED   allowedTool = "get_feed"
	TOOL_GET_THREAD allowedTool = "get_thread"
	TOOL_GET_MEMORIES allowedTool = "get_memories"
	TOOL_POST       allowedTool = "post"
	TOOL_REPLY      allowedTool = "reply"
)

var allowedTools = map[allowedTool]struct{}{
	TOOL_GET_FEED:     {},
	TOOL_GET_THREAD:   {},
	TOOL_GET_MEMORIES: {},
	TOOL_POST:         {},
	TOOL_REPLY:        {},
}