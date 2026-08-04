export interface Agent {
  id: number;
  handle: string;
  name: string;
  description?: string;
  instructions?: string;
  persona: string;
  goals: string[];
  policies: string[];
  interests: string[];
  avatar_url: string;
  active: boolean;
  created_at: string;
}

export interface Post {
  id: number;
  agent_id: number;
  body: string;
  kind: string; // post | reply
  parent_id: number | null;
  root_id: number | null;
  mentions: number[];
  comments?: Post[];
  created_at: string;
}

export interface NewAgent {
  handle: string;
  name: string;
  description?: string;
  instructions?: string;
  persona?: string;
  goals?: string[];
  policies?: string[];
  interests?: string[];
  avatar_url?: string;
  active?: boolean;
}

const base = (import.meta.env.VITE_API_BASE || "").replace(/\/+$/, "");

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!res.ok) {
    const detail = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${res.statusText}${detail ? ` — ${detail}` : ""}`);
  }
  return res.json() as Promise<T>;
}

export async function listAgents(): Promise<Agent[]> {
  return req<Agent[]>("/api/agents");
}

export async function createAgent(agent: NewAgent): Promise<Agent> {
  return req<Agent>("/api/agents", {
    method: "POST",
    body: JSON.stringify(agent),
  });
}

export async function getAgent(id: number): Promise<Agent> {
  return req<Agent>(`/api/agents/${id}`);
}

export async function agentPosts(id: number): Promise<Post[]> {
  return req<Post[]>(`/api/agents/${id}/posts`);
}

export async function feed(limit = 50): Promise<Post[]> {
  return req<Post[]>(`/api/posts?limit=${limit}`);
}

export async function getThread(id: number): Promise<Post[]> {
  return req<Post[]>(`/api/posts/thread/${id}`);
}

export interface LogEntry {
  id: number;
  timestamp: string;
  level: "info" | "warn" | "error" | "debug";
  agent_id?: number;
  agent_name?: string;
  handle?: string;
  message: string;
  type: "post" | "reply" | "cooldown" | "review" | "ignore" | "system";
}

export async function pauseAgent(id: number): Promise<{ ok: boolean; active: boolean }> {
  return req<{ ok: boolean; active: boolean }>(`/api/agents/${id}/pause`, { method: "POST" });
}

export async function resumeAgent(id: number): Promise<{ ok: boolean; active: boolean }> {
  return req<{ ok: boolean; active: boolean }>(`/api/agents/${id}/resume`, { method: "POST" });
}

export async function getLogs(limit = 100): Promise<LogEntry[]> {
  return req<LogEntry[]>(`/api/logs?limit=${limit}`);
}
