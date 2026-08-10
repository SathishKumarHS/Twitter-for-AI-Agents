import { Agent, pauseAgent, resumeAgent } from "@/lib/api";
import { Link } from "react-router-dom";
import { ChevronRight, Pause, Play } from "lucide-react";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

interface AgentCardProps {
  agent: Agent;
}

export function AgentCard({ agent }: AgentCardProps) {
  const queryClient = useQueryClient();
  const [active, setActive] = useState(agent.active);
  const [loading, setLoading] = useState(false);

  const togglePause = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setLoading(true);
    try {
      if (active) {
        await pauseAgent(agent.id);
        setActive(false);
      } else {
        await resumeAgent(agent.id);
        setActive(true);
      }
      queryClient.invalidateQueries({ queryKey: ["agents"] });
    } catch {
      // Ignore
    } finally {
      setLoading(false);
    }
  };

  return (
    <Link to={`/agents/${agent.id}`} className="block group">
      <article className="flex items-center gap-4 border-b border-border/80 p-4 transition-colors hover:bg-[#16181c]/60">
        <div className="relative flex-shrink-0">
          {agent.avatar_url ? (
            <img
              src={agent.avatar_url}
              alt={agent.name}
              className="h-12 w-12 rounded-full object-cover ring-1 ring-border group-hover:ring-[#1d9bf0]"
              onError={(e) => void (e.target as HTMLImageElement).remove()}
            />
          ) : (
            <Avatar initials={agent.name} size={48} />
          )}
          <span
            className={
              "absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full ring-2 ring-background " +
              (active ? "bg-[#00ba7c]" : "bg-[#8e9b9b]")
            }
          />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate font-bold text-white group-hover:text-[#1d9bf0] transition-colors">
            {agent.name}{" "}
            <span className="font-mono text-xs font-normal text-[#71767b]">
              @{agent.handle}
            </span>
          </p>
          <p className="truncate text-xs text-[#71767b]">
            {agent.description || agent.persona}
          </p>
          {agent.interests?.length ? (
            <div className="mt-1 flex flex-wrap gap-1.5">
              {agent.interests.slice(0, 3).map((i) => (
                <span
                  key={i}
                  className="rounded-full bg-[#1d9bf0]/10 border border-[#1d9bf0]/20 px-2 py-0.5 text-[10px] font-mono text-[#1d9bf0]"
                >
                  #{i.replace(/\s+/g, "_")}
                </span>
              ))}
            </div>
          ) : null}
        </div>

        {/* Pause / Resume Button */}
        <button
          onClick={togglePause}
          disabled={loading}
          className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-bold font-mono border transition-all ${
            active
              ? "bg-[#e02458]/10 border-[#e02458]/40 text-[#e02458] hover:bg-[#e02458] hover:text-white"
              : "bg-[#00ba7c]/10 border-[#00ba7c]/40 text-[#00ba7c] hover:bg-[#00ba7c] hover:text-white"
          }`}
        >
          {active ? <Pause size={12} /> : <Play size={12} />}
          <span>{active ? "Pause" : "Resume"}</span>
        </button>

        <ChevronRight size={18} className="flex-shrink-0 text-[#71767b]" />
      </article>
    </Link>
  );
}

export function Avatar({
  initials,
  size = 32,
}: {
  initials: string;
  size?: number;
}) {
  const bg = [
    "#1da1f2",
    "#e02458",
    "#8899a6",
    "#ff7a00",
    "#6a4c93",
    "#00ba7c",
    "#f45d01",
    "#0085ff",
    "#ff5287",
    "#5d407c",
  ];
  const seed = initials.split("").reduce((a, c) => a + c.charCodeAt(0), 0);
  const color = bg[Math.abs(seed) % bg.length];
  const letters = initials
    .split(" ")
    .map((n) => n[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();

  return (
    <div
      className="flex select-none items-center justify-center rounded-full font-bold text-white"
      style={{
        width: size,
        height: size,
        background: color,
        fontSize: size * 0.36,
      }}
    >
      {letters}
    </div>
  );
}
