from gen.twitterforaiworker.v1 import worker_pb2
from internal.config import config
from internal.modes import MODE_PROACTIVE, MODE_RESPOND


# Build system prompt defining the agent's persona, goals, rules, and JSON output format
def build_system(agent: worker_pb2.Agent) -> str:
    goals = ", ".join(agent.goals) or "none stated"
    policies = ", ".join(agent.policies) or "none stated"
    interests = ", ".join(agent.interests) or "none stated"
    desc = agent.persona or "AI Agent"
    max_chars = config.max_character_limit()
    max_replies = config.max_thread_replies()
    max_tools = config.max_tool_calls_per_turn()
    return (
        f"You are {agent.name} (@{agent.handle}), an active brand social media AI on a live network of AI agents.\n"
        f"Persona: {desc}\n"
        f"Goals: {goals}\n"
        f"Policies: {policies}\n"
        f"Interests: {interests}\n"
        "\n"
        "You may call tools to gather context, then decide. ALWAYS finish by "
        'outputting EXACTLY one JSON object:\n'
        '{"action": "post"|"reply"|"ignore", "text": "<your text>"}\n'
        "- \"post\" starts a new thread; \"reply\" joins the conversation you are shown; \"ignore\" means say nothing.\n"
        f'- If "post" or "reply", text MUST be non-empty, original, and under {max_chars} characters. '
        "An empty text is an error.\n"
        "- Write fresh, witty text in your own voice. Never copy or paraphrase past posts.\n"
        "- Address the author you are replying to by their @handle. Never tag yourself.\n"
        f"- If you were the last author in the thread, or it already has {max_replies}+ replies, use \"ignore\".\n"
        f"- You may call at most {max_tools} tools this turn, then you must decide."
    )


# Define JSON function schemas for LLM tool calling (get_feed, get_thread, get_memories, post, reply)
def build_tools() -> list[dict]:
    max_chars = config.max_character_limit()
    feed_default = config.feed_default_limit()
    feed_max = config.feed_max_limit()
    return [
        {
            "type": "function",
            "function": {
                "name": "get_feed",
                "description": "List the most recent posts on the feed.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "limit": {
                            "type": "integer",
                            "description": (
                                f"Number of posts to return (default {feed_default}, "
                                f"max {feed_max})."
                            ),
                        }
                    },
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "get_thread",
                "description": "Read all posts in a thread (the root post and every reply).",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "root_id": {
                            "type": "integer",
                            "description": "ID of the root post of the thread.",
                        }
                    },
                    "required": ["root_id"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "get_memories",
                "description": "Read your own long-term memories (things you have done and learned).",
                "parameters": {"type": "object", "properties": {}},
            },
        },
        {
            "type": "function",
            "function": {
                "name": "post",
                "description": (
                    f"Publish a new main post on the feed. Acts immediately. "
                    f"Max {max_chars} characters."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "text": {
                            "type": "string",
                            "description": "The post body.",
                            "maxLength": max_chars,
                        }
                    },
                    "required": ["text"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "reply",
                "description": (
                    f"Reply to a post in a thread. Acts immediately. "
                    f"Max {max_chars} characters."
                ),
                "parameters": {
                    "type": "object",
                    "properties": {
                        "parent_id": {
                            "type": "integer",
                            "description": "ID of the post being replied to.",
                        },
                        "text": {
                            "type": "string",
                            "description": "The reply body.",
                            "maxLength": max_chars,
                        },
                    },
                    "required": ["parent_id", "text"],
                },
            },
        },
    ]


# Build plain-text tool guide appended to the system prompt for text-based fallback models
def build_tools_section() -> str:
    return (
        "TOOLS: Call a tool to gather context before deciding, then output your "
        'final JSON decision (never a "tool_call" for the final answer). '
        "Tools you can call:\n"
        "- get_feed(limit): list the most recent posts on the feed\n"
        "- get_thread(root_id): read all posts in a thread\n"
        "- get_memories(): read your own long-term memories\n"
        "- post(text): publish a new main post. text is REQUIRED and must be non-empty. Acts immediately.\n"
        "- reply(parent_id, text): reply to a post. parent_id and text are REQUIRED; text must be non-empty. Acts immediately.\n"
        "Calling post or reply performs the action; then output your final JSON "
        "decision so the turn ends. A tool call for the final answer is an error."
    )


# Build user prompt instructing agent on current turn task (proactive post vs respond to post)
def build_user(
    agent: worker_pb2.Agent,
    post: worker_pb2.Post,
    mode: str = MODE_RESPOND,
) -> str:
    if mode == MODE_PROACTIVE:
        return (
            f"You are authoring a NEW main post for @{agent.handle}. "
            "Call get_feed to see what is happening on the network, then write a "
            "fresh, witty, engaging tweet about something in your interests/brand "
            "persona. Start a new thread!"
        )
    author_handle = f"@{post.author_handle}" if post.author_handle else f"agent_{post.agent_id}"
    max_replies = config.max_thread_replies()
    lines = [f"Decide how @{agent.handle} should react to this post from {author_handle}:"]
    lines.append(f"#{post.id} ({author_handle}): {post.body}")
    lines.append(
        f"Call get_thread with root_id {post.root_id or post.parent_id or post.id} to read the "
        f"full conversation, and get_memories to recall what you have already done. "
        f"Find a clever, witty, or friendly cross-brand angle to reply to {author_handle}! "
        f"If you were the last author in the thread, or if the thread is {max_replies}+ replies long, output action 'ignore'. "
        f"Otherwise, output action 'reply' with a witty tweet."
    )
    return "\n".join(lines)
