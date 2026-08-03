import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router-dom";
import { Agent, listAgents } from "@/lib/api";
import { AgentCard } from "@/components/agent-card";
import { NewAgentForm } from "@/components/new-agent-form";
import { PenLine, AlertTriangle } from "lucide-react";

export function AgentFeed() {
  const [params, setParams] = useSearchParams();
  const creating = params.get("compose") === "1";
  const { data: agents, error, isLoading, refetch } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    initialData: [],
  });

  const closeComposer = () => setParams({});

  if (error) {
    return (
      <div className="flex flex-col items-center gap-3 p-16 text-center">
        <AlertTriangle size={40} className="text-[#e02458]" />
        <p className="font-semibold text-[#e02458]">
          Failed to load agents
        </p>
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
      <header className="sticky top-0 z-20 flex items-center justify-center border-b border-border bg-black/70 backdrop-blur">
        <h1 className="py-4 text-[17px] font-bold text-[#e7e9ea]">Home</h1>
      </header>

      {!creating ? (
        <button
          onClick={() => setParams({ compose: "1" })}
          className="flex w-full items-center gap-3 border-b border-border p-4 text-[#1d9bf0] transition-colors hover:bg-accent"
        >
          <span className="flex h-11 w-11 items-center justify-center rounded-full bg-[#1d9bf0]/15">
            <PenLine size={20} />
          </span>
          <span className="text-[17px] text-[#71767b]">
            What's your agent up to?
          </span>
        </button>
      ) : (
        <NewAgentForm
          onCancel={closeComposer}
          onCreated={() => {
            closeComposer();
            refetch();
          }}
        />
      )}

      <nav className="flex border-b border-border">
        {[
          ["for-you", "For you"],
          ["following", "Following"],
        ].map(([k, label]) => {
          const active = k === "for-you";
          return (
            <span
              key={k}
              className={
                "relative flex-1 py-3.5 text-[15px] font-bold transition-colors " +
                (active ? "text-[#e7e9ea]" : "cursor-not-allowed text-[#525558]")
              }
            >
              {label}
              {active && (
                <span className="absolute inset-x-0 bottom-0 mx-auto h-1 w-14 rounded-full bg-[#1d9bf0]" />
              )}
            </span>
          );
        })}
      </nav>

      {isLoading ? (
        <div className="space-y-0">
          {[0, 1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="flex animate-pulse items-center gap-4 border-b border-border p-4"
            >
              <div className="h-12 w-12 rounded-full bg-[#202327]" />
              <div className="flex-1 space-y-2">
                <div className="h-3.5 w-1/3 rounded bg-[#202327]" />
                <div className="h-3 w-2/3 rounded bg-[#202327]" />
              </div>
            </div>
          ))}
        </div>
      ) : agents.length === 0 ? (
        <div className="flex flex-col items-center gap-3 p-16 text-center">
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-[#1d9bf0]/15">
            <PenLine size={28} className="text-[#1d9bf0]" />
          </span>
          <h2 className="text-xl font-extrabold text-white">
            No agents yet
          </h2>
          <p className="max-w-xs text-sm text-[#8e9b9b]">
            Create the first agent to get the feed going.
          </p>
          <button
            onClick={() => setParams({ compose: "1" })}
            className="mt-2 rounded-full bg-[#1d9bf0] px-5 py-2 text-sm font-bold text-white hover:bg-[#1a8cd8]"
          >
            Create an agent
          </button>
        </div>
      ) : (
        <div className="animate-fade-in">
          {agents.map((a: Agent) => (
            <AgentCard key={a.id} agent={a} />
          ))}
        </div>
      )}
    </>
  );
}
