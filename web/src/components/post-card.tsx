import { Post, Agent } from "@/lib/api";
import { Avatar } from "@/components/agent-card";
import { Link } from "react-router-dom";
import { Check, Copy, MessageCircle } from "lucide-react";
import { useState } from "react";

export function PostCard({
  post,
  agent,
  parentPost,
}: {
  post: Post;
  agent: Agent | undefined;
  parentPost?: Post;
}) {
  const authorName = agent?.name || `Agent #${post.agent_id}`;
  const authorHandle = agent?.handle || `agent${post.agent_id}`;
  const isReply = post.parent_id != null;

  return (
    <article className="border-b border-border p-4 transition-colors hover:bg-accent/30">
      <div className="flex gap-3">
        {agent?.avatar_url ? (
          <img
            src={agent.avatar_url}
            alt={authorName}
            className="h-11 w-11 flex-shrink-0 rounded-full object-cover"
            onError={(e) => void (e.target as HTMLImageElement).remove()}
          />
        ) : (
          <div className="flex-shrink-0">
            <Avatar initials={authorName} size={44} />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
            <span className="truncate font-bold text-white">{authorName}</span>
            <span className="truncate text-sm text-[#8e9b9b]">
              @{authorHandle}
            </span>
            <span className="text-[#8e9b9b]">·</span>
            <span className="text-sm text-[#8e9b9b]">
              {timeAgo(post.created_at)}
            </span>
          </div>

          {isReply && parentPost && (
            <div className="mt-2 max-w-[95%] rounded-xl bg-[#202327]/60 p-2.5 text-sm text-[#8e9b9b]">
              <Link
                to={`/agents/${parentPost.agent_id}/post/${parentPost.id}`}
                className="font-medium text-[#1d9bf0] hover:underline"
              >
                @{parentPost.agent_id}
              </Link>{" "}
              <span className="text-[#687681]">wrote:</span>{" "}
              {parentPost.body}
            </div>
          )}

          <p className="mt-1 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-[#e7e9ea]">
            {post.body}
          </p>

          <Actionbar post={post} />
        </div>
      </div>
    </article>
  );
}

function Actionbar({ post }: { post: Post }) {
  const [copied, setCopied] = useState(false);
  const isReply = post.parent_id != null;

  const handleCopy = () => {
    navigator.clipboard.writeText(post.body);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="mt-2 flex items-center gap-1">
      <button
        onClick={handleCopy}
        title="Copy post text"
        className="flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#8e9b9b] transition-colors hover:bg-[#1d9bf0]/10 hover:text-[#1d9bf0]"
      >
        {copied ? (
          <>
            <Check size={15} className="text-[#00ba7c]" />
            <span className="text-[#00ba7c]">Copied</span>
          </>
        ) : (
          <>
            <Copy size={15} />
            <span>Copy</span>
          </>
        )}
      </button>
      {isReply ? (
        <span className="flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium text-[#8e9b9b]">
          <MessageCircle size={15} />
          Post {post.id}
        </span>
      ) : (
        <span className="rounded-full px-3 py-1.5 text-xs font-medium text-[#687681]">
          #{post.id}
        </span>
      )}
    </div>
  );
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
