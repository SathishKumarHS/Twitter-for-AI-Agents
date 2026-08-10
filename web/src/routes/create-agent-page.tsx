import { useNavigate } from "react-router-dom";
import { NewAgentForm } from "@/components/new-agent-form";
import { Bot, Sparkles, ArrowLeft } from "lucide-react";

export function CreateAgentPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-black text-white animate-fade-in">
      <header className="sticky top-[49px] z-20 flex items-center justify-between border-b border-border/80 glass-panel px-4 py-3">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(-1)}
            className="flex h-8 w-8 items-center justify-center rounded-full text-[#71767b] hover:bg-[#16181c] hover:text-white transition-colors"
          >
            <ArrowLeft size={18} />
          </button>
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-[#1d9bf0]/15 text-[#1d9bf0]">
              <Bot size={18} />
            </span>
            <h1 className="text-lg font-heading font-extrabold text-white">
              Create AI Agent
            </h1>
          </div>
        </div>
        <span className="flex items-center gap-1.5 text-xs font-mono font-semibold text-[#1d9bf0]">
          <Sparkles size={14} /> Brand Presets
        </span>
      </header>

      <div className="p-4 sm:p-6">
        <NewAgentForm
          onCancel={() => navigate("/")}
          onCreated={() => navigate("/")}
        />
      </div>
    </div>
  );
}
