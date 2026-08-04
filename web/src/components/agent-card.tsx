import { Agent } from "@/lib/api";
import { Link } from "react-router-dom";
import { ChevronRight } from "lucide-react";

interface AgentCardProps {
  agent: Agent;
}

export function AgentCard({ agent }: AgentCardProps) {
  return (
    <Link to={`/agents/${agent.id}`} className="block">
      <article className="flex items-center gap-4 border-b border-border p-4 transition-colors hover:bg-accent">
        <div className="relative flex-shrink-0">
          <Avatar initials={agent.name} size={48} />
          <span
            className={
              "absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full ring-2 ring-background " +
              (agent.active ? "bg-[#00ba7c]" : "bg-[#8e9b9b]")
            }
          />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate font-bold text-white">
            {agent.name}{" "}
            <span className="font-normal text-[#687681]">@{agent.handle}</span>
          </p>
          <p className="truncate text-sm text-[#8e9b9b]">
            {agent.description || agent.persona}
          </p>
          {agent.interests?.length ? (
            <div className="mt-1 flex flex-wrap gap-1.5">
              {agent.interests.slice(0, 3).map((i) => (
                <span
                  key={i}
                  className="rounded-full bg-[#1d9bf0]/10 px-2 py-0.5 text-[11px] font-medium text-[#1d9bf0]"
                >
                  #{i.replace(/\s+/g, "_")}
                </span>
              ))}
            </div>
          ) : null}
        </div>
        <ChevronRight size={18} className="flex-shrink-0 text-[#687681]" />
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
    "#1da1f2", "#e02458", "#8899a6", "#ff7a00", "#6a4c93",
    "#00ba7c", "#f45d01", "#0085ff", "#ff5287", "#5d407c",
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
