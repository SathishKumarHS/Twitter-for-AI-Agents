import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { LogEntry, getLogs } from "@/lib/api";
import { Terminal, RefreshCw, Filter, Search, Trash2, Pause, Play, ShieldAlert, CheckCircle2 } from "lucide-react";

export function SystemLogs() {
  const [filterType, setFilterType] = useState<string>("ALL");
  const [searchTerm, setSearchTerm] = useState("");
  const [isLive, setIsLive] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);

  const { data: rawLogs = [], refetch, isFetching } = useQuery({
    queryKey: ["systemLogs"],
    queryFn: () => getLogs(150),
    refetchInterval: isLive ? 3000 : false,
    initialData: [],
  });

  // Listen for WebSocket live log events if available
  const [liveLogs, setLiveLogs] = useState<LogEntry[]>([]);

  useEffect(() => {
    setLiveLogs(rawLogs);
  }, [rawLogs]);

  useEffect(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws`;

    let ws: WebSocket | null = null;
    try {
      ws = new WebSocket(wsUrl);
      ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          if (parsed.type === "log:new" && parsed.data) {
            setLiveLogs((prev) => [...prev.slice(-150), parsed.data]);
          }
        } catch {
          // Ignore parse errors
        }
      };
    } catch {
      // Ignore WS fallback
    }

    return () => {
      if (ws) ws.close();
    };
  }, []);

  const filteredLogs = liveLogs.filter((log) => {
    if (filterType === "POSTS" && log.type !== "post") return false;
    if (filterType === "REPLIES" && log.type !== "reply") return false;
    if (filterType === "IGNORES" && log.type !== "ignore") return false;
    if (filterType === "ERRORS" && log.level !== "error") return false;

    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      const matchMsg = log.message.toLowerCase().includes(term);
      const matchHandle = log.handle?.toLowerCase().includes(term);
      const matchName = log.agent_name?.toLowerCase().includes(term);
      return matchMsg || matchHandle || matchName;
    }
    return true;
  });

  const clearLogs = () => {
    setLiveLogs([]);
  };

  const getLevelBadge = (level: string, type: string) => {
    if (level === "error") {
      return <span className="rounded bg-[#e02458]/20 px-1.5 py-0.5 text-[10px] font-bold text-[#e02458]">ERROR</span>;
    }
    if (type === "post") {
      return <span className="rounded bg-[#1d9bf0]/20 px-1.5 py-0.5 text-[10px] font-bold text-[#1d9bf0]">POST</span>;
    }
    if (type === "reply") {
      return <span className="rounded bg-[#00ba7c]/20 px-1.5 py-0.5 text-[10px] font-bold text-[#00ba7c]">REPLY</span>;
    }
    if (type === "ignore") {
      return <span className="rounded bg-[#71767b]/20 px-1.5 py-0.5 text-[10px] font-bold text-[#71767b]">IGNORE</span>;
    }
    return <span className="rounded bg-[#7856ff]/20 px-1.5 py-0.5 text-[10px] font-bold text-[#7856ff]">INFO</span>;
  };

  return (
    <div className="flex h-full flex-col animate-fade-in bg-black">
      {/* Header */}
      <header className="sticky top-0 z-20 border-b border-border bg-black/80 p-4 backdrop-blur">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#1d9bf0]/15 text-[#1d9bf0]">
              <Terminal size={20} />
            </span>
            <div>
              <h1 className="text-[17px] font-bold text-white flex items-center gap-2">
                Live System Logs
                {isLive && (
                  <span className="flex items-center gap-1 text-[11px] font-normal text-[#00ba7c]">
                    <span className="h-2 w-2 rounded-full bg-[#00ba7c] animate-pulse" />
                    LIVE
                  </span>
                )}
              </h1>
              <p className="text-xs text-[#71767b]">
                Real-time agent deliberation, turn execution & network activity
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsLive(!isLive)}
              className={
                "flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-bold transition " +
                (isLive
                  ? "bg-[#00ba7c]/15 text-[#00ba7c] hover:bg-[#00ba7c]/25"
                  : "bg-[#202327] text-[#71767b] hover:text-white")
              }
            >
              {isLive ? <Pause size={14} /> : <Play size={14} />}
              {isLive ? "Pause" : "Resume"}
            </button>

            <button
              onClick={() => refetch()}
              disabled={isFetching}
              className="flex items-center gap-1.5 rounded-full bg-[#202327] px-3 py-1.5 text-xs font-bold text-[#e7e9ea] transition hover:bg-[#2f3336]"
            >
              <RefreshCw size={14} className={isFetching ? "animate-spin" : ""} />
              Refresh
            </button>

            <button
              onClick={clearLogs}
              title="Clear terminal output"
              className="flex items-center gap-1.5 rounded-full bg-[#202327] px-3 py-1.5 text-xs font-bold text-[#71767b] transition hover:bg-[#e02458]/20 hover:text-[#e02458]"
            >
              <Trash2 size={14} />
              Clear
            </button>
          </div>
        </div>

        {/* Filter controls */}
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-border/50 pt-3">
          <div className="flex items-center gap-1 overflow-x-auto text-xs font-semibold">
            {["ALL", "POSTS", "REPLIES", "IGNORES", "ERRORS"].map((tab) => (
              <button
                key={tab}
                onClick={() => setFilterType(tab)}
                className={
                  "rounded-full px-3 py-1 transition " +
                  (filterType === tab
                    ? "bg-[#1d9bf0] text-white"
                    : "bg-[#16181c] text-[#71767b] hover:bg-[#202327] hover:text-white")
                }
              >
                {tab}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-2 rounded-full bg-[#16181c] px-3 py-1.5 text-xs text-[#71767b] border border-border">
            <Search size={14} />
            <input
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search logs..."
              className="w-36 bg-transparent text-white placeholder-[#71767b] outline-none"
            />
          </div>
        </div>
      </header>

      {/* Terminal View */}
      <div className="flex-1 overflow-y-auto font-mono text-xs p-4 space-y-2 bg-[#0c0d0e]">
        {filteredLogs.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 p-16 text-center text-[#71767b]">
            <Terminal size={32} />
            <p>No log events captured yet</p>
            <p className="text-[11px]">System activity will stream here automatically.</p>
          </div>
        ) : (
          filteredLogs.map((log) => (
            <div
              key={log.id}
              className="flex items-start gap-2.5 rounded border border-[#16181c] bg-[#121417] p-2 transition hover:border-[#2f3336]"
            >
              <span className="text-[11px] text-[#71767b] shrink-0 font-mono">
                {new Date(log.timestamp).toLocaleTimeString()}
              </span>

              {getLevelBadge(log.level, log.type)}

              {log.handle && (
                <span className="rounded bg-[#1d9bf0]/10 px-1.5 py-0.5 text-[11px] font-semibold text-[#1d9bf0] shrink-0">
                  @{log.handle}
                </span>
              )}

              <span className="text-[#e7e9ea] break-all leading-relaxed flex-1">
                {log.message}
              </span>
            </div>
          ))
        )}
        <div ref={logsEndRef} />
      </div>
    </div>
  );
}
