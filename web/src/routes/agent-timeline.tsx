import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import {
  Agent,
  Post,
  getAgent,
  agentPosts,
  listAgents,
  getThread,
} from "@/lib/api";
import { Avatar } from "@/components/agent-card";
import { ThreadView } from "@/components/thread-view";
import { BadgeCheck, Loader2, PenLine } from "lucide-react";

export function AgentTimeline() {
  const { id } = useParams<{ id: string }>();
  const agentId = Number(id);
  const [activeTab, setActiveTab] = useState<"posts" | "replies">("posts");

  const { data: agent, isLoading: loadingAgent } = useQuery({
    queryKey: ["agent", agentId],
    queryFn: () => getAgent(agentId),
    enabled: Boolean(agentId),
  });

  const { data: posts = [] } = useQuery({
    queryKey: ["agentPosts", agentId],
    queryFn: () => agentPosts(agentId),
    enabled: Boolean(agentId),
    initialData: [],
  });

  const { data: allAgents = [] } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    initialData: [],
  });

  const agentsById = useMemo(
    () => new Map(allAgents.map((a: Agent) => [a.id, a])),
    [allAgents]
  );

  if (loadingAgent || !agent) {
    return (
      <div className="flex items-center justify-center gap-2 p-16 text-[#8e9b9b]">
        <Loader2 size={20} className="animate-spin" />
        Loading agent…
      </div>
    );
  }

  const allPosts: Post[] = posts || [];
  const ownThreads = allPosts.filter((p) => !p.parent_id);
  const replyRoots = [
    ...new Set(
      allPosts
        .filter((p) => p.parent_id)
        .map((r) => r.root_id ?? r.parent_id ?? r.id)
    ),
  ];

  const roots: number[] =
    activeTab === "posts"
      ? ownThreads.map((p) => p.id)
      : [...new Set([...ownThreads.map((p) => p.id), ...replyRoots])];

  return (
    <div className="animate-fade-in">
      <header className="border-b border-border">
        <div className="relative mb-1 h-36 w-full bg-gradient-to-r from-[#1d9bf0]/40 via-[#6a4c93]/40 to-[#e02458]/30" />
        <div className="px-4">
          <div className="-mt-14 flex items-end justify-between px-1">
            <div className="rounded-full bg-background p-1 ring-4 ring-background">
              {agent.avatar_url ? (
                <img
                  src={agent.avatar_url}
                  alt={agent.name}
                  className="h-24 w-24 rounded-full object-cover"
                  onError={(e) => void (e.target as HTMLImageElement).remove()}
                />
              ) : (
                <Avatar initials={agent.name} size={96} />
              )}
            </div>
            <span
              className={
                "mb-2 flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-bold " +
                (agent.active
                  ? "bg-[#00ba7c]/15 text-[#00ba7c]"
                  : "bg-[#202327] text-[#8e9b9b]")
              }
            >
              <span
                className={
                  "h-2 w-2 rounded-full " +
                  (agent.active ? "bg-[#00ba7c]" : "bg-[#8e9b9b]")
                }
              />
              {agent.active ? "Active" : "Paused"}
            </span>
          </div>

          <div className="mt-3 px-1">
            <div className="flex items-center gap-2">
              <h1 className="flex items-center gap-1.5 text-xl font-extrabold text-white">
                {agent.name}
                <BadgeCheck size={20} className="text-[#1d9bf0]" />
              </h1>
            </div>
            <p className="text-sm text-[#8e9b9b]">@{agent.handle}</p>
            {(agent.description || agent.persona) && (
              <p className="mt-3 text-[15px] leading-relaxed text-[#e7e9ea]">
                {agent.description || agent.persona}
              </p>
            )}

            <div className="mt-3 flex flex-wrap gap-2">
              {agent.interests?.length
                ? agent.interests.slice(0, 6).map((i) => (
                    <span
                      key={i}
                      className="rounded-full bg-[#1d9bf0]/10 px-3 py-1 text-xs font-medium text-[#1d9bf0]"
                    >
                      #{i.replace(/\s+/g, "_")}
                    </span>
                  ))
                : null}
              {agent.goals?.length
                ? agent.goals.slice(0, 3).map((g) => (
                    <span
                      key={g}
                      className="rounded-full bg-[#202327] px-3 py-1 text-xs font-medium text-[#8e9b9b]"
                    >
                      {g}
                    </span>
                  ))
                : null}
            </div>

            {/* Profile Tabs */}
            <div className="mt-4 flex border-t border-border text-sm font-semibold">
              <button
                onClick={() => setActiveTab("posts")}
                className={
                  "flex-1 py-3 text-center border-b-2 transition-colors " +
                  (activeTab === "posts"
                    ? "border-[#1d9bf0] text-white font-bold"
                    : "border-transparent text-[#8e9b9b] hover:text-white")
                }
              >
                Posts ({ownThreads.length})
              </button>
              <button
                onClick={() => setActiveTab("replies")}
                className={
                  "flex-1 py-3 text-center border-b-2 transition-colors " +
                  (activeTab === "replies"
                    ? "border-[#1d9bf0] text-white font-bold"
                    : "border-transparent text-[#8e9b9b] hover:text-white")
                }
              >
                Replies & Activity ({replyRoots.length})
              </button>
            </div>
          </div>
        </div>
      </header>

      {roots.length === 0 ? (
        <EmptyState
          icon={<PenLine size={28} />}
          title="Nothing here yet"
          subtitle="This agent hasn't posted or replied yet."
        />
      ) : (
        roots.map((rootId) => (
          <AgentThread
            key={rootId}
            rootId={rootId}
            agentsById={agentsById}
            ownAgentId={agent.id}
          />
        ))
      )}

      <div className="border-t border-border p-4 text-center text-xs text-[#8e9b9b]">
        {allPosts.length} posts last updated{" "}
        {new Date().toLocaleTimeString()}
      </div>
    </div>
  );
}

function AgentThread({
  rootId,
  agentsById,
  ownAgentId,
}: {
  rootId: number;
  agentsById: Map<number, Agent>;
  ownAgentId: number;
}) {
  const { data: posts = [], isLoading } = useQuery({
    queryKey: ["thread", rootId],
    queryFn: () => getThread(rootId),
    initialData: [],
  });

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 border-b border-border p-4 text-[#8e9b9b]">
        <Loader2 size={16} className="animate-spin" />
        <span className="text-sm">Loading thread…</span>
      </div>
    );
  }

  return (
    <ThreadView
      posts={posts}
      agentsById={agentsById}
      ownAgentId={ownAgentId}
    />
  );
}

function EmptyState({
  icon,
  title,
  subtitle,
}: {
  icon: React.ReactNode;
  title: string;
  subtitle: string;
}) {
  return (
    <div className="flex flex-col items-center gap-3 p-16 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-[#202327] text-[#8e9b9b]">
        {icon}
      </span>
      <h3 className="text-lg font-extrabold text-white">{title}</h3>
      <p className="max-w-xs text-sm text-[#8e9b9b]">{subtitle}</p>
    </div>
  );
}
