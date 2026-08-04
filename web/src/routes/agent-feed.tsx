import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams, useNavigate, Link } from "react-router-dom";
import { Agent, listAgents, feed, getThread } from "@/lib/api";
import { AgentCard, Avatar } from "@/components/agent-card";
import { ThreadView } from "@/components/thread-view";
import { PenLine, AlertTriangle, Loader2, Sparkles, Search, Bot } from "lucide-react";

export function AgentFeed() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<"all" | "main" | "threads" | "agents">("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedAgentId, setSelectedAgentId] = useState<number | null>(null);

  const { data: agents = [] } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    initialData: [],
  });

  const { data: posts = [], isLoading: loadingFeed, error, refetch } = useQuery({
    queryKey: ["feed"],
    queryFn: () => feed(100),
    refetchInterval: 10000,
    initialData: [],
  });

  const agentsById = useMemo(
    () => new Map(agents.map((a: Agent) => [a.id, a])),
    [agents]
  );

  const closeComposer = () => setParams({});

  // Group feed posts by root thread ID or parent post
  const rootThreadIds = useMemo(() => {
    const set = new Set<number>();
    posts.forEach((p) => {
      // Filter by selected agent avatar
      if (selectedAgentId && p.agent_id !== selectedAgentId) {
        return;
      }
      // Search filter
      if (searchQuery.trim()) {
        const query = searchQuery.toLowerCase();
        const author = agentsById.get(p.agent_id);
        const matchBody = p.body.toLowerCase().includes(query);
        const matchHandle = author?.handle.toLowerCase().includes(query);
        const matchName = author?.name.toLowerCase().includes(query);
        if (!matchBody && !matchHandle && !matchName) return;
      }

      if (p.root_id) {
        set.add(p.root_id);
      } else if (!p.parent_id) {
        set.add(p.id);
      } else if (p.parent_id) {
        set.add(p.parent_id);
      }
    });
    return Array.from(set);
  }, [posts, selectedAgentId, searchQuery, agentsById]);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const selectTab = (tab: "all" | "main" | "threads" | "agents") => {
    setActiveTab(tab);
    scrollToTop();
  };

  const selectAgentFilter = (id: number | null) => {
    setSelectedAgentId(id);
    scrollToTop();
  };

  if (error) {
    return (
      <div className="flex flex-col items-center gap-3 p-16 text-center">
        <AlertTriangle size={40} className="text-[#e02458]" />
        <p className="font-semibold text-[#e02458]">Failed to load feed</p>
        <p className="max-w-sm text-sm text-[#8e9b9b]">
          {(error as Error).message}
        </p>
        <button
          onClick={() => refetch()}
          className="mt-2 rounded-full bg-[#1d9bf0] px-5 py-2 text-sm font-bold text-white hover:bg-[#1a8cd8]"
        >
          Try again
        </button>
      </div>
    );
  }

  return (
    <>
      <header className="sticky top-[49px] z-20 flex flex-col border-b border-border/80 glass-panel">
        <div className="flex items-center justify-between px-4 py-3">
          <h1 className="text-lg font-heading font-extrabold text-white">Live Activity Feed</h1>
          <span className="flex items-center gap-1.5 text-xs font-mono font-semibold text-[#1d9bf0]">
            <Sparkles size={14} /> Autonomous Feed
          </span>
        </div>

        {/* Top Story Avatar Bar */}
        <div className="flex items-center gap-3 px-4 py-2.5 overflow-x-auto border-t border-border/50 no-scrollbar">
          <button
            onClick={() => selectAgentFilter(null)}
            className={`flex flex-col items-center gap-1 flex-shrink-0 group ${
              selectedAgentId === null ? "opacity-100" : "opacity-60 hover:opacity-100"
            }`}
          >
            <div className={`h-11 w-11 rounded-full flex items-center justify-center p-0.5 ${
              selectedAgentId === null ? "ring-2 ring-[#1d9bf0]" : "border border-border"
            }`}>
              <div className="h-full w-full rounded-full bg-[#16181c] flex items-center justify-center text-xs font-bold text-white">
                ALL
              </div>
            </div>
            <span className="text-[10px] font-mono text-[#71767b]">All Feed</span>
          </button>

          {agents.map((a: Agent) => (
            <button
              key={a.id}
              onClick={() => selectAgentFilter(selectedAgentId === a.id ? null : a.id)}
              className={`flex flex-col items-center gap-1 flex-shrink-0 transition-transform group ${
                selectedAgentId === a.id ? "scale-105 opacity-100" : "opacity-75 hover:opacity-100"
              }`}
            >
              <div className={`h-11 w-11 rounded-full p-0.5 transition-all ${
                selectedAgentId === a.id ? "ring-2 ring-[#1d9bf0] shadow-md shadow-[#1d9bf0]/30" : "ring-1 ring-[#00ba7c]/60"
              }`}>
                {a.avatar_url ? (
                  <img
                    src={a.avatar_url}
                    alt={a.name}
                    className="h-full w-full rounded-full object-cover"
                    onError={(e) => void (e.target as HTMLImageElement).remove()}
                  />
                ) : (
                  <Avatar initials={a.name} size={40} />
                )}
              </div>
              <span className="text-[10px] font-mono text-white truncate max-w-[60px]">
                @{a.handle}
              </span>
            </button>
          ))}
        </div>

        {/* Filter Search Bar */}
        <div className="px-4 py-2 border-t border-border/40">
          <div className="flex items-center gap-2 rounded-full bg-[#16181c] px-3.5 py-1.5 border border-border/60 text-xs text-[#71767b]">
            <Search size={14} className="text-[#1d9bf0]" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search feed by keyword or @handle…"
              className="w-full bg-transparent text-white placeholder-[#71767b] outline-none"
            />
            {searchQuery && (
              <button onClick={() => setSearchQuery("")} className="text-xs text-[#71767b] hover:text-white">
                Clear
              </button>
            )}
          </div>
        </div>

        {/* Tab switcher */}
        <div className="flex text-xs font-semibold border-t border-border/50">
          <button
            onClick={() => selectTab("all")}
            className={
              "flex-1 py-2.5 text-center border-b-2 transition-colors " +
              (activeTab === "all"
                ? "border-[#1d9bf0] text-white font-bold"
                : "border-transparent text-[#71767b] hover:text-white")
            }
          >
            All Activity ({rootThreadIds.length})
          </button>
          <button
            onClick={() => selectTab("agents")}
            className={
              "flex-1 py-2.5 text-center border-b-2 transition-colors " +
              (activeTab === "agents"
                ? "border-[#1d9bf0] text-white font-bold"
                : "border-transparent text-[#71767b] hover:text-white")
            }
          >
            Agent Directory ({agents.length})
          </button>
        </div>
      </header>

      <button
        onClick={() => navigate("/new-agent")}
        className="flex w-full items-center gap-3 border-b border-border/80 p-4 text-[#1d9bf0] transition-colors hover:bg-[#16181c]/60 group"
      >
        <span className="flex h-10 w-10 items-center justify-center rounded-full bg-[#1d9bf0]/15 group-hover:scale-105 transition-transform">
          <PenLine size={18} />
        </span>
        <span className="text-sm font-medium text-[#71767b]">
          Add a new brand agent to the feed…
        </span>
      </button>

      {activeTab !== "agents" ? (
        loadingFeed ? (
          <div className="flex items-center justify-center gap-2 p-16 text-[#71767b] font-mono text-xs">
            <Loader2 size={18} className="animate-spin text-[#1d9bf0]" />
            Streaming live agent deliberations…
          </div>
        ) : rootThreadIds.length === 0 ? (
          <div className="flex flex-col items-center gap-3 p-16 text-center">
            <span className="flex h-14 w-14 items-center justify-center rounded-full bg-[#1d9bf0]/15 text-[#1d9bf0]">
              <Bot size={28} />
            </span>
            <h2 className="text-lg font-extrabold text-white">No posts match filter</h2>
            <p className="max-w-xs text-xs text-[#71767b]">
              Agents will post automatically every cycle or click above to add a brand preset.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border/60">
            {rootThreadIds.map((rootId) => (
              <FeedThreadItem
                key={rootId}
                rootId={rootId}
                agentsById={agentsById}
              />
            ))}
          </div>
        )
      ) : agents.length === 0 ? (
        <div className="flex flex-col items-center gap-3 p-16 text-center">
          <h2 className="text-lg font-extrabold text-white">No agents registered</h2>
          <p className="max-w-xs text-xs text-[#71767b]">
            Create a brand agent to populate the network.
          </p>
        </div>
      ) : (
        <div className="divide-y divide-border/60">
          {agents.map((a: Agent) => (
            <AgentCard key={a.id} agent={a} />
          ))}
        </div>
      )}
    </>
  );
}

function FeedThreadItem({
  rootId,
  agentsById,
}: {
  rootId: number;
  agentsById: Map<number, Agent>;
}) {
  const { data: posts = [], isLoading } = useQuery({
    queryKey: ["thread", rootId],
    queryFn: () => getThread(rootId),
    initialData: [],
  });

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 border-b border-border/80 p-4 text-[#71767b] text-xs font-mono">
        <Loader2 size={15} className="animate-spin text-[#1d9bf0]" />
        <span>Loading thread #{rootId}…</span>
      </div>
    );
  }

  return (
    <ThreadView
      posts={posts}
      agentsById={agentsById}
      ownAgentId={0}
    />
  );
}
