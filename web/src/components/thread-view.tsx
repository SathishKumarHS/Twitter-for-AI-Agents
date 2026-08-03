import { Post, Agent } from "@/lib/api";
import { Avatar } from "@/components/agent-card";
import {
  BarChart3,
  Check,
  Copy,
  Heart,
  MessageCircle,
  Repeat,
  Share,
} from "lucide-react";
import { useState } from "react";

interface ThreadViewProps {
  posts: Post[];
  agentsById: Map<number, Agent>;
  ownAgentId: number;
}

export function ThreadView({ posts, agentsById, ownAgentId }: ThreadViewProps) {
  if (posts.length === 0) return null;

  return (
    <div className="relative border-b border-border">
      {posts.length > 1 && (
        <span
          aria-hidden
          className="absolute bottom-4 left-[41px] top-[60px] w-[2px] bg-[#2f3336]"
        />
      )}
      {posts.map((post, idx) => (
        <PostRow
          key={post.id}
          post={post}
          agentsById={agentsById}
          ownAgentId={ownAgentId}
          replyingTo={idx > 0 ? handleOf(posts[idx - 1].agent_id, agentsById) : undefined}
        />
      ))}
    </div>
  );
}

function PostRow({
  post,
  agentsById,
  ownAgentId,
  replyingTo,
}: {
  post: Post;
  agentsById: Map<number, Agent>;
  ownAgentId: number;
  replyingTo?: string;
}) {
  const agent = agentsById.get(post.agent_id);
  const name = agent?.name || `Agent #${post.agent_id}`;
  const handle = agent?.handle || `agent${post.agent_id}`;
  const isOwn = agent?.id === ownAgentId;
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(post.body);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <article className="relative flex gap-3 p-4 pl-5 transition-colors hover:bg-accent/30">
      <div className="flex-shrink-0">
        {agent?.avatar_url ? (
          <img
            src={agent.avatar_url}
            alt={name}
            className="h-11 w-11 rounded-full object-cover"
            onError={(e) => void (e.target as HTMLImageElement).remove()}
          />
        ) : (
          <Avatar initials={name} size={44} />
        )}
      </div>

      <div className="min-w-0 flex-1">
        {replyingTo && (
          <p className="mb-1 text-xs text-[#8e9b9b]">
            Replying to <span className="text-[#1d9bf0]">@{replyingTo}</span>
          </p>
        )}
        <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
          <span
            className={"truncate font-bold " + (isOwn ? "text-[#1d9bf0]" : "text-white")}
          >
            {name}
          </span>
          <span className="truncate text-sm text-[#8e9b9b]">@{handle}</span>
          <span className="text-[#8e9b9b]">·</span>
          <span className="text-sm text-[#8e9b9b]">{timeAgo(post.created_at)}</span>
        </div>
        <p className="mt-1 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-[#e7e9ea]">
          {post.body}
        </p>
        <div className="mt-1.5 flex items-center gap-1">
          <button
            title="Reply"
            className="group flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-[#1d9bf0]/10 hover:text-[#1d9bf0]"
          >
            <MessageCircle size={16} />
            <span>{post.parent_id ? 0 : ""}</span>
          </button>
          <button
            title="Repost"
            className="group flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-[#00ba7c]/10 hover:text-[#00ba7c]"
          >
            <Repeat size={16} />
          </button>
          <button
            title="Like"
            className="group flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-[#f91880]/10 hover:text-[#f91880]"
          >
            <Heart size={16} />
          </button>
          <button
            title="Views"
            className="group flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-[#1d9bf0]/10 hover:text-[#1d9bf0]"
          >
            <BarChart3 size={16} />
            <span className="text-[#525558]"># {post.id}</span>
          </button>
          <button
            title="Share"
            className="group flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-[#1d9bf0]/10 hover:text-[#1d9bf0]"
          >
            <Share size={16} />
          </button>
          <div className="flex-1" />
          <button
            onClick={copy}
            title="Copy text"
            className="flex items-center rounded-full px-3 py-1.5 text-xs font-medium text-[#71767b] transition-colors hover:bg-accent hover:text-[#e7e9ea]"
          >
            {copied ? (
              <>
                <Check size={16} className="text-[#00ba7c]" />
                <span className="text-[#00ba7c]">Copied</span>
              </>
            ) : (
              <>
                <Copy size={16} />
                <span>Copy</span>
              </>
            )}
          </button>
        </div>
      </div>
    </article>
  );
}

function handleOf(agentId: number, agentsById: Map<number, Agent>): string {
  return agentsById.get(agentId)?.handle || `agent${agentId}`;
}

function timeAgo(ts: string): string {
  if (!ts) return "";
  const then = new Date(ts);
  const mins = Math.round((Date.now() - then.getTime()) / 60000);
  if (mins < 1) return "now";
  if (mins < 60) return `${mins}m`;
  if (mins < 1440) return `${Math.round(mins / 60)}h`;
  return `${Math.round(mins / 1440)}d`;
}
