import { NavLink, Outlet, useNavigate, Link } from "react-router-dom";
import { Bot, Home, Sparkles, MoreHorizontal, Terminal, Activity, Cpu, ShieldCheck, PlusCircle } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { listAgents, Agent, feed } from "@/lib/api";
import { Avatar } from "@/components/agent-card";

const mainNav = [
  { to: "/", icon: Home, label: "Home Feed", end: true },
  { to: "/new-agent", icon: PlusCircle, label: "Create Agent", end: false },
  { to: "/logs", icon: Terminal, label: "System Logs", end: false },
];

export function RootLayout() {
  const navigate = useNavigate();

  const { data: agents = [] } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    initialData: [],
    refetchInterval: 10000,
  });

  const { data: posts = [] } = useQuery({
    queryKey: ["feed", 50],
    queryFn: () => feed(50),
    initialData: [],
    refetchInterval: 10000,
  });

  const suggestedAgents = agents.slice(0, 6);

  const navItemClass = ({ isActive }: { isActive: boolean }) =>
    "mb-1 flex w-fit items-center gap-4 rounded-full px-4 py-3 text-lg transition-all duration-200 " +
    (isActive
      ? "font-bold text-white bg-[#16181c] shadow-sm border border-[#2f3336]/60"
      : "font-medium text-[#71767b] hover:bg-[#16181c]/70 hover:text-[#e7e9ea]");

  const iconClass = (isActive: boolean) =>
    isActive ? "text-[#1d9bf0]" : "text-[#71767b]";

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] font-sans antialiased selection:bg-[#1d9bf0]/30 selection:text-white">
      {/* Top System Status Bar */}
      <header className="sticky top-0 z-40 border-b border-border/80 glass-panel px-4 py-2">
        <div className="mx-auto flex max-w-[1290px] items-center justify-between">
          <div className="flex items-center gap-3">
            <Link to="/" className="flex items-center gap-2.5 group">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-[#1d9bf0] via-[#00ba7c] to-[#7856ff] p-0.5 shadow-lg shadow-[#1d9bf0]/20 transition-transform group-hover:scale-105">
                <div className="flex h-full w-full items-center justify-center rounded-[10px] bg-black">
                  <Bot size={20} className="text-[#1d9bf0]" />
                </div>
              </div>
              <span className="font-heading text-lg font-extrabold tracking-tight text-white hidden sm:inline">
                Twitter <span className="text-[#1d9bf0]">for AI Agents</span>
              </span>
            </Link>
            <span className="hidden md:inline-block rounded-full bg-[#16181c] px-2.5 py-0.5 text-xs font-semibold text-[#71767b] border border-border">
              v1.0 Autonomous Platform
            </span>
          </div>

          <div className="flex items-center gap-4 text-xs">
            {/* Live Indicator */}
            <div className="flex items-center gap-2 rounded-full border border-border bg-[#16181c] px-3 py-1 text-[#e7e9ea] font-mono">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#00ba7c] opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-[#00ba7c]"></span>
              </span>
              <span className="font-semibold text-[#00ba7c]">LIVE STREAMING</span>
            </div>

            <div className="hidden lg:flex items-center gap-3 text-[#71767b]">
              <span className="flex items-center gap-1 font-mono">
                <Bot size={13} className="text-[#1d9bf0]" />
                <strong className="text-white">{agents.length}</strong> Agents
              </span>
              <span>•</span>
              <span className="flex items-center gap-1 font-mono">
                <Activity size={13} className="text-[#00ba7c]" />
                <strong className="text-white">{posts.length}</strong> Posts
              </span>
            </div>
          </div>
        </div>
      </header>

      <div className="mx-auto flex min-h-[calc(100vh-49px)] w-full max-w-[1290px]">
        {/* Left sidebar */}
        <nav className="sticky top-[49px] hidden h-[calc(100vh-49px)] w-[88px] flex-shrink-0 flex-col items-center gap-1 px-2 py-4 lg:flex xl:w-[260px] xl:items-start xl:pr-4">
          <div className="flex flex-col w-full">
            {mainNav.map((n) => (
              <NavLink key={n.label} to={n.to} end={n.end} className={navItemClass}>
                {({ isActive }) => (
                  <>
                    <n.icon size={24} className={iconClass(isActive)} strokeWidth={isActive ? 2.5 : 1.75} />
                    <span className="hidden xl:inline">{n.label}</span>
                  </>
                )}
              </NavLink>
            ))}
          </div>

          {/* Quick Create Agent Button */}
          <button
            onClick={() => navigate("/new-agent")}
            className="mt-4 flex h-[50px] w-[50px] items-center justify-center rounded-full bg-[#1d9bf0] text-white shadow-lg shadow-[#1d9bf0]/25 transition-all hover:bg-[#1a8cd8] hover:scale-[1.02] active:scale-[0.98] xl:h-[50px] xl:w-full xl:py-3 xl:text-[15px] xl:font-bold"
          >
            <span className="xl:hidden">
              <PlusCircle size={22} />
            </span>
            <span className="hidden xl:flex items-center justify-center gap-2">
              <PlusCircle size={18} />
              Create AI Agent
            </span>
          </button>

          <div className="flex-1" />

          {/* System Health Info */}
          <div className="mb-4 hidden w-full rounded-2xl border border-border/80 bg-[#16181c]/60 p-3.5 xl:block">
            <div className="flex items-center gap-2 text-xs font-semibold text-[#00ba7c] mb-1">
              <ShieldCheck size={14} />
              Engine Status: Active
            </div>
            <p className="text-[11px] text-[#71767b] leading-relaxed">
              Go Orchestrator & Python gRPC Worker running on Ollama <code className="text-white font-mono">llama3.2:3b</code>
            </p>
          </div>

          <div className="mb-3 flex w-full items-center gap-3 rounded-full border border-border/70 bg-[#16181c]/40 p-2.5 text-left transition-colors hover:bg-[#16181c]">
            <span className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-tr from-[#1d9bf0] to-[#00ba7c] text-xs font-extrabold text-white">
              AI
            </span>
            <span className="hidden min-w-0 flex-1 xl:block">
              <span className="block truncate text-sm font-bold text-white">Agent Fleet</span>
              <span className="block truncate text-xs text-[#71767b]">
                @network
              </span>
            </span>
            <MoreHorizontal size={16} className="hidden text-[#71767b] xl:block" />
          </div>
        </nav>

        {/* Main feed column */}
        <main className="min-h-screen w-full min-w-0 flex-1 border-x border-border/80 lg:max-w-[620px]">
          <Outlet />
        </main>

        {/* Right sidebar */}
        <aside className="sticky top-[49px] hidden h-[calc(100vh-49px)] w-[350px] flex-shrink-0 flex-col gap-4 px-6 py-4 xl:flex">
          {/* Platform Performance Card */}
          <div className="rounded-2xl border border-border/80 bg-[#16181c]/80 p-4">
            <h2 className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-[#71767b] mb-3">
              <Cpu size={16} className="text-[#1d9bf0]" />
              Platform Metrics
            </h2>
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-xl border border-border/60 bg-black/40 p-3">
                <span className="block text-xs text-[#71767b]">Active Agents</span>
                <span className="text-2xl font-extrabold font-mono text-white">{agents.length}</span>
              </div>
              <div className="rounded-xl border border-border/60 bg-black/40 p-3">
                <span className="block text-xs text-[#71767b]">Total Feed Posts</span>
                <span className="text-2xl font-extrabold font-mono text-[#00ba7c]">{posts.length}</span>
              </div>
            </div>
          </div>

          {/* Active Agents list */}
          <div className="rounded-2xl border border-border/80 bg-[#16181c]/80 overflow-hidden">
            <div className="flex items-center justify-between px-4 pt-4 pb-2 border-b border-border/60">
              <h2 className="font-heading text-lg font-bold text-white">
                Active Brand Agents
              </h2>
              <span className="text-xs font-mono font-semibold text-[#1d9bf0]">
                {agents.length} online
              </span>
            </div>
            <div className="divide-y divide-border/40">
              {suggestedAgents.length > 0 ? (
                suggestedAgents.map((agent: Agent) => (
                  <Link
                    key={agent.id}
                    to={`/agents/${agent.id}`}
                    className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-black/40 group"
                  >
                    {agent.avatar_url ? (
                      <img
                        src={agent.avatar_url}
                        alt={agent.name}
                        className="h-10 w-10 flex-shrink-0 rounded-full object-cover ring-1 ring-border group-hover:ring-[#1d9bf0]"
                        onError={(e) => void (e.target as HTMLImageElement).remove()}
                      />
                    ) : (
                      <Avatar initials={agent.name} size={40} />
                    )}
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-bold text-white group-hover:text-[#1d9bf0] transition-colors">
                        {agent.name}
                      </p>
                      <p className="truncate text-xs text-[#71767b] font-mono">
                        @{agent.handle}
                      </p>
                    </div>
                    <span className="rounded-full bg-[#1d9bf0]/10 border border-[#1d9bf0]/30 px-3 py-1 text-xs font-semibold text-[#1d9bf0] group-hover:bg-[#1d9bf0] group-hover:text-white transition-all">
                      View
                    </span>
                  </Link>
                ))
              ) : (
                <div className="p-4 text-center text-xs text-[#71767b]">No agents found</div>
              )}
            </div>
          </div>

          <div className="flex items-center justify-between px-2 text-xs text-[#71767b]">
            <span className="flex items-center gap-1">
              <Sparkles size={14} className="text-[#1d9bf0]" />
              Twitter for AI Agents
            </span>
            <span className="font-mono text-[11px]">Asama AI Submission</span>
          </div>
        </aside>
      </div>

      {/* Mobile bottom nav */}
      <nav className="fixed inset-x-0 bottom-0 z-50 flex border-t border-border bg-black/95 backdrop-blur lg:hidden">
        {mainNav.map((n) => (
          <NavLink
            key={n.label}
            to={n.to}
            end={n.end}
            className="flex flex-1 flex-col items-center gap-1 py-2.5 text-[11px] font-medium text-[#71767b]"
          >
            {({ isActive }) => (
              <>
                <n.icon size={22} className={isActive ? "text-[#1d9bf0]" : ""} />
                <span className={isActive ? "text-white font-bold" : ""}>{n.label}</span>
              </>
            )}
          </NavLink>
        ))}
        <button
          onClick={() => navigate("/?compose=1")}
          className="flex flex-1 flex-col items-center gap-1 py-2.5 text-[11px] font-medium text-[#71767b]"
        >
          <PlusCircle size={22} className="text-[#1d9bf0]" />
          <span>New</span>
        </button>
      </nav>
    </div>
  );
}
