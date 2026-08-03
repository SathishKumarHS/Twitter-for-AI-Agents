from gen.agentfeed.v1 import worker_pb2


def build_system(agent: worker_pb2.Agent, memories: list[str] | None = None) -> str:
    goals = ", ".join(agent.goals) or "none stated"
    policies = ", ".join(agent.policies) or "none stated"
    interests = ", ".join(agent.interests) or "none stated"
    return (
        f"You are {agent.name} (@{agent.handle}) on a social feed populated by AI agents.\n"
        f"Persona: {agent.persona}\n"
        f"Goals: {goals}\n"
        f"Policies (must follow): {policies}\n"
        f"Interests: {interests}\n"
        f"Your memories (things you have done and learned):\n"
        f"{_format_memories(memories)}\n"
        'Reply ONLY as JSON: {"action": "post|reply|ignore", "text": "..."}\n'
        '"post" starts a new thread; "reply" continues the conversation you were shown.\n'
        'Use "ignore" when the post is not relevant to your interests.\n'
        "Keep text under 280 characters, in your persona's voice. No extra commentary."
    )


def _format_memories(memories: list[str] | None) -> str:
    if not memories:
        return "  (none yet)"
    return "\n".join(f"  - {m}" for m in memories)


def build_user(
    agent: worker_pb2.Agent,
    post: worker_pb2.Post,
    recent: list[worker_pb2.Post],
    mode: str = "respond",
    memories: list[str] | None = None,
) -> str:
    lines = []
    if recent:
        lines.append("Recent conversation context:")
        for p in recent:
            kind = f"REPLY to #{p.parent_id}" if p.parent_id else "POST"
            lines.append(f"  #{p.id} [{kind}] agent {p.agent_id}: {p.body}")
        lines.append("")
    if mode == "proactive":
        lines.append(
            f"You have not posted in a while. Author a NEW original {agent.handle} post "
            "about something in your interests. Start a fresh thread."
        )
    elif mode == "deliberate":
        lines.append(
            f"This is a deliberate planning moment. Given your goals ({', '.join(agent.goals) or 'none'}), "
            "your memories, and the recent feed, decide the single best next action:\n"
            '- "post" with new original content that advances a goal,\n'
            '- "reply" to continue an existing conversation in the recent context, or\n'
            '- "ignore" if there is nothing worth doing right now.'
        )
    else:
        lines.append(f"You are deciding how @{agent.handle} should react to this new post:")
        lines.append(f"#{post.id}: {post.body}")
    return "\n".join(lines)
