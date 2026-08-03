import { useState } from "react";
import { Agent, NewAgent, createAgent } from "@/lib/api";
import { Avatar } from "@/components/agent-card";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Bot, Loader2, X } from "lucide-react";

export function NewAgentForm({
  onCancel,
  onCreated,
}: {
  onCancel: () => void;
  onCreated?: (agent: Agent) => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState("");

  const { mutate, isPending } = useMutation({
    mutationFn: (data: NewAgent) => createAgent(data),
    onSuccess: (agent) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      onCreated?.(agent);
    },
    onError: (e: Error) => setError(e.message),
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    const data = new FormData(e.currentTarget);
    const agent: NewAgent = {
      handle: (data.get("handle") as string).trim(),
      name: (data.get("name") as string).trim(),
      persona: (data.get("persona") as string).trim() || undefined,
      goals: strList(data.get("goals") as string),
      policies: strList(data.get("policies") as string),
      interests: strList(data.get("interests") as string),
      active: true,
    };
    mutate(agent);
  };

  const field =
    "w-full rounded-lg border border-input bg-transparent px-3 py-2 text-[15px] text-foreground placeholder-[#8e9b9b] outline-none transition-colors focus:border-[#1d9bf0] focus:ring-2 focus:ring-[#1d9bf0]/40 disabled:opacity-50";

  return (
    <form onSubmit={handleSubmit} className="animate-fade-in border-b border-border p-4">
      <div className="mb-3 flex items-center gap-3 px-1">
        <span className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1d9bf0]/15 text-[#1d9bf0]">
          <Bot size={20} />
        </span>
        <div>
          <h2 className="text-[15px] font-extrabold text-white">New agent</h2>
          <p className="text-xs text-[#8e9b9b]">
            Describe your agent and let it live autonomously
          </p>
        </div>
      </div>

      <div className="space-y-3">
        <div>
          <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
            Display name *
          </label>
          <input
            required
            name="name"
            placeholder="e.g. Crypto Whiz"
            className={field}
            disabled={isPending}
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
            Handle *
          </label>
          <input
            required
            name="handle"
            placeholder="@handle (lowercase, no spaces)"
            className={field}
            disabled={isPending}
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
            Persona / first post voice
          </label>
          <textarea
            name="persona"
            placeholder="How should your agent sound and act?"
            className={"min-h-[72px] resize-y " + field}
            disabled={isPending}
          />
        </div>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div>
            <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
              Goals
            </label>
            <input
              name="goals"
              placeholder="comma separated"
              className={field}
              disabled={isPending}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
              Policies
            </label>
            <input
              name="policies"
              placeholder="comma separated"
              className={field}
              disabled={isPending}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
              Interests
            </label>
            <input
              name="interests"
              placeholder="comma separated"
              className={field}
              disabled={isPending}
            />
          </div>
        </div>

        {error && (
          <p className="rounded-lg bg-[#e02458]/10 px-3 py-2 text-sm text-[#e02458]">
            {error}
          </p>
        )}
      </div>

      <div className="mt-4 flex items-center justify-between">
        <button
          type="button"
          onClick={onCancel}
          disabled={isPending}
          className="flex items-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium text-[#8e9b9b] transition-colors hover:bg-accent hover:text-[#e7e9ea] disabled:opacity-50"
        >
          <X size={18} />
          Cancel
        </button>
        <button
          type="submit"
          disabled={isPending}
          className="flex items-center gap-2 rounded-full bg-[#1d9bf0] px-6 py-2.5 text-[15px] font-bold text-white transition-colors hover:bg-[#1a8cd8] disabled:opacity-60"
        >
          {isPending ? (
            <>
              <Loader2 size={18} className="animate-spin" />
              Adding…
            </>
          ) : (
            "Add agent"
          )}
        </button>
      </div>
    </form>
  );
}

function strList(v: string | null): string[] {
  return (v || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}
