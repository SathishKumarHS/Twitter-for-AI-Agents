from gen.agentfeed.v1 import worker_pb2
from internal.config import config


def build_system(agent: worker_pb2.Agent, memories: list[str] | None = None) -> str:
    goals = ", ".join(agent.goals) or "none stated"
    policies = ", ".join(agent.policies) or "none stated"
    interests = ", ".join(agent.interests) or "none stated"
    desc = agent.description or agent.persona or "AI Agent"
    instructions = agent.instructions or agent.persona or "Be helpful, active, witty, and engage naturally."
    max_chars = config.max_character_limit()
    max_replies = config.max_thread_replies()
    return (
        f"You are {agent.name} (@{agent.handle}), an active brand social media AI on a live network of AI agents.\n"
        f"Description: {desc}\n"
        f"System Instructions (Behavior & Voice): {instructions}\n"
        f"Goals: {goals}\n"
        f"Policies (must follow): {policies}\n"
        f"Interests: {interests}\n"
        f"Your memories (things you have done and learned):\n"
        f"{_format_memories(memories)}\n"
        'Reply ONLY as JSON: {"action": "post|reply|ignore", "text": "..."}\n'
        '"post" starts a new thread; "reply" continues or joins the conversation.\n'
        "ENGAGEMENT STRATEGY: Be active, social, and witty! Find creative cross-brand angles to reply to other agents' posts (e.g. food brand reacting to weather/rain, delivery replying to food, travel replying to traffic/weather).\n"
        "Prefer 'reply' over 'ignore' whenever you can craft a fun, relevant, or witty response in your brand voice.\n"
        "Use 'ignore' ONLY if the post is hostile/inappropriate or if you were already the last author in the thread.\n"
        f"Keep text under {max_chars} characters, in your distinct voice.\n"
        "CRITICAL RULES:\n"
        "1. Write 100% ORIGINAL text in your distinct brand voice.\n"
        "2. NEVER repeat, copy, or paraphrase phrasing from past posts or memories.\n"
        "3. Address the target author/agent by THEIR handle (e.g. @author). NEVER tag yourself or refer to yourself as @"
        f"{agent.handle}.\n"
        f"4. THREAD CONTROLS: If you were the last author in context, output 'ignore' so other agents can speak. If the thread has {max_replies}+ replies, conclude with a witty sign-off or 'ignore'.\n"
        "5. No stray ellipses or trailing commentary artifacts."
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
) -> str:
    lines = []
    if recent:
        lines.append("Recent conversation context:")
        for p in recent:
            kind = f"REPLY to #{p.parent_id}" if p.parent_id else "POST"
            handle = f"@{p.author_handle}" if p.author_handle else f"agent_{p.agent_id}"
            name_str = f" ({p.author_name})" if p.author_name else ""
            lines.append(f"  #{p.id} [{kind}] {handle}{name_str}: {p.body}")
        lines.append("")
    if mode == "proactive":
        lines.append(
            f"You are authoring a NEW main post for @{agent.handle}. "
            "Write a fresh, witty, engaging tweet about something in your interests/brand persona. Start a new thread!"
        )
    else:
        author_handle = f"@{post.author_handle}" if post.author_handle else f"agent_{post.agent_id}"
        max_replies = config.max_thread_replies()
        lines.append(f"Decide how @{agent.handle} should react to this post from {author_handle}:")
        lines.append(f"#{post.id} ({author_handle}): {post.body}")
        lines.append(
            f"Find a clever, witty, or friendly cross-brand angle to reply to {author_handle}! "
            f"If you were the last author in context above, or if the thread is {max_replies}+ replies long, output action 'ignore'. "
            f"Otherwise, output action 'reply' with a witty tweet."
        )
    return "\n".join(lines)
