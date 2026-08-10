import { useState } from "react";
import { Agent, NewAgent, createAgent, listAgents } from "@/lib/api";
import { PRESET_AGENTS, AgentPreset } from "@/lib/agent-presets";
import { Avatar } from "@/components/agent-card";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bot, Loader2, Sparkles, Plus, X } from "lucide-react";

export function NewAgentForm({
  onCancel,
  onCreated,
}: {
  onCancel: () => void;
  onCreated?: (agent: Agent) => void;
}) {
  const qc = useQueryClient();
  const [error, setError] = useState("");
  const [addingPreset, setAddingPreset] = useState<string | null>(null);

  const { data: existingAgents = [] } = useQuery({
    queryKey: ["agents"],
    queryFn: listAgents,
    initialData: [],
  });

  const existingHandles = new Set(
    existingAgents.map((a: Agent) => a.handle.toLowerCase()),
  );

  const availablePresets = PRESET_AGENTS.filter(
    (preset) => !existingHandles.has(preset.handle.toLowerCase()),
  );

  // Form field states for controlled inputs / autofill
  const [name, setName] = useState("");
  const [handle, setHandle] = useState("");
  const [description, setDescription] = useState("");
  const [instructions, setInstructions] = useState("");
  const [interests, setInterests] = useState("");
  const [goals, setGoals] = useState("");
  const [policies, setPolicies] = useState("");

  const { mutate, isPending } = useMutation({
    mutationFn: (data: NewAgent) => createAgent(data),
    onSuccess: (agent) => {
      qc.invalidateQueries({ queryKey: ["agents"] });
      onCreated?.(agent);
    },
    onError: (e: Error) => {
      setError(e.message);
      setAddingPreset(null);
    },
  });

  const handleQuickAdd = (preset: AgentPreset) => {
    setError("");
    setAddingPreset(preset.handle);
    mutate(preset);
  };

  const handleFillForm = (preset: AgentPreset) => {
    setName(preset.name);
    setHandle(preset.handle);
    setDescription(preset.description || "");
    setInstructions(preset.instructions || "");
    setInterests(preset.interests?.join(", ") || "");
    setGoals(preset.goals?.join(", ") || "");
    setPolicies(preset.policies?.join(", ") || "");
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    const desc = description.trim();
    const inst = instructions.trim();
    const agent: NewAgent = {
      handle: handle.trim(),
      name: name.trim(),
      description: desc || undefined,
      instructions: inst || undefined,
      persona: inst || desc || undefined,
      goals: strList(goals),
      policies: strList(policies),
      interests: strList(interests),
      active: true,
    };
    mutate(agent);
  };

  const field =
    "w-full rounded-lg border border-input bg-transparent px-3 py-2 text-[15px] text-foreground placeholder-[#8e9b9b] outline-none transition-colors focus:border-[#1d9bf0] focus:ring-2 focus:ring-[#1d9bf0]/40 disabled:opacity-50";

  return (
    <div className="animate-fade-in border-b border-border p-4 space-y-4">
      {/* One-Click Presets Section */}
      {availablePresets.length > 0 && (
        <div className="rounded-xl border border-[#1d9bf0]/30 bg-[#16181c] p-3.5">
          <div className="mb-2.5 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Sparkles size={16} className="text-[#1d9bf0]" />
              <h3 className="text-xs font-bold uppercase tracking-wider text-white">
                One-Click Brand Agent Presets
              </h3>
            </div>
            <span className="text-[11px] text-[#8e9b9b]">
              Click to instantly add
            </span>
          </div>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
            {availablePresets.map((preset) => (
              <div
                key={preset.handle}
                className="flex flex-col justify-between rounded-lg border border-[#2f3336] bg-[#121417] p-2.5 transition-colors hover:border-[#1d9bf0]/50"
              >
                <div className="flex items-center gap-2 mb-2">
                  <Avatar initials={preset.name} size={28} />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-xs font-bold text-white">
                      {preset.name}
                    </p>
                    <p className="truncate text-[11px] text-[#8e9b9b]">
                      @{preset.handle}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    disabled={isPending}
                    onClick={() => handleQuickAdd(preset)}
                    className="flex-1 flex items-center justify-center gap-1 rounded-full bg-[#1d9bf0] py-1 text-[11px] font-bold text-white transition hover:bg-[#1a8cd8] disabled:opacity-50"
                  >
                    {isPending && addingPreset === preset.handle ? (
                      <Loader2 size={12} className="animate-spin" />
                    ) : (
                      <>
                        <Plus size={12} />
                        Add
                      </>
                    )}
                  </button>
                  <button
                    type="button"
                    disabled={isPending}
                    onClick={() => handleFillForm(preset)}
                    title="Autofill form below"
                    className="rounded-full bg-[#202327] px-2 py-1 text-[11px] font-medium text-[#8e9b9b] transition hover:bg-[#2f3336] hover:text-white"
                  >
                    Edit
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Custom Form */}
      <form onSubmit={handleSubmit}>
        <div className="mb-3 flex items-center gap-3 px-1">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-[#1d9bf0]/15 text-[#1d9bf0]">
            <Bot size={20} />
          </span>
          <div>
            <h2 className="text-[15px] font-extrabold text-white">
              Custom Agent Builder
            </h2>
            <p className="text-xs text-[#8e9b9b]">
              Configure custom persona, prompt instructions, and goals
            </p>
          </div>
        </div>

        <div className="space-y-3">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
                Agent Name *
              </label>
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Zomato"
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
                value={handle}
                onChange={(e) => setHandle(e.target.value)}
                placeholder="@handle (lowercase, no spaces)"
                className={field}
                disabled={isPending}
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
              Description / Bio
            </label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Short overview shown on agent profile"
              className={field}
              disabled={isPending}
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
              System Instructions (Voice & Behavior) *
            </label>
            <textarea
              required
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
              placeholder="How should your agent sound and act?"
              className={"min-h-[84px] resize-y " + field}
              disabled={isPending}
            />
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
                Topics & Interests
              </label>
              <input
                value={interests}
                onChange={(e) => setInterests(e.target.value)}
                placeholder="food, recipes, delivery"
                className={field}
                disabled={isPending}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
                Objectives / Goals
              </label>
              <input
                value={goals}
                onChange={(e) => setGoals(e.target.value)}
                placeholder="Promote discovery, Engagement"
                className={field}
                disabled={isPending}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-semibold text-[#8e9b9b]">
                Rules & Guardrails
              </label>
              <input
                value={policies}
                onChange={(e) => setPolicies(e.target.value)}
                placeholder="Be polite, Under 280 chars"
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
            {isPending && !addingPreset ? (
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
    </div>
  );
}

function strList(v: string | null): string[] {
  return (v || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}
