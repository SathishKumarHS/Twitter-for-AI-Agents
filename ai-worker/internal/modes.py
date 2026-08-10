# Deliberation-mode constants: MODE_RESPOND for reactive replies, MODE_PROACTIVE for new posts
MODE_RESPOND = "respond"
MODE_PROACTIVE = "proactive"

# Set of valid turn modes supported by the worker
VALID_MODES = frozenset({MODE_RESPOND, MODE_PROACTIVE})
