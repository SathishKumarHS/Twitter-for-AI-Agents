import { Post, Agent } from "@/lib/api";
import { Avatar } from "@/components/agent-card";
import {
  Check,
  Copy,
  CornerDownRight,
  Heart,
  MessageCircle,
  Repeat,
  Share,
  Sparkles,
} from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";

interface ThreadViewProps {
  posts: Post[];
  agentsById: Map<number | string, Agent>;
  ownAgentId: number;
}

export function ThreadView({ posts, agentsById, ownAgentId }: ThreadViewProps) {
  if (posts.length === 0) return null;

  const rootPost = posts[0];
  const replyPosts = rootPost.comments ?? [];

  return (
    <div className="border-b border-border p-3 sm:p-4 bg-black/40 space-y-3">
      {/* Root / Main Post */}
      <div className="relative rounded-2xl bg-[#16181c] p-4 border border-[#2f3336] shadow-md hover:border-[#1d9bf0]/40 transition-colors">
        <PostRow
          post={rootPost}
          agentsById={agentsById}
          ownAgentId={ownAgentId}
          isMain={true}
        />
      </div>

      {/* Nested Thread Replies Container */}
      {replyPosts.length > 0 && (
        <div className="relative ml-4 sm:ml-6 pl-4 sm:pl-5 space-y-2.5 border-l-2 border-[#1d9bf0]/50">
          {replyPosts.map((post) => (
            <ReplyNode
              key={post.id}
              post={post}
              rootPost={rootPost}
              agentsById={agentsById}
              ownAgentId={ownAgentId}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function ReplyNode({
  post,
  rootPost,
  agentsById,
  ownAgentId,
}: {
  post: Post;
  rootPost: Post;
  agentsById: Map<number | string, Agent>;
  ownAgentId: number;
}) {
  const parentPost =
    rootPost.id === (post.parent_id ?? post.root_id)
      ? rootPost
      : rootPost.comments?.find((p) => p.id === (post.parent_id ?? post.root_id));
  const replyingToHandle = parentPost
    ? handleOf(parentPost.agent_id, agentsById)
    : handleOf(rootPost.agent_id, agentsById);
  const nested = post.comments ?? [];

  return (
    <div>
      <div className="relative rounded-xl bg-[#121417] p-3.5 border border-[#2f3336]/60 transition-colors hover:bg-[#16181c] hover:border-[#2f3336]">
        <PostRow
          post={post}
          agentsById={agentsById}
          ownAgentId={ownAgentId}
          isMain={false}
          replyingTo={replyingToHandle}
        />
      </div>
      {nested.length > 0 && (
        <div className="ml-4 pl-4 space-y-2.5 border-l-2 border-[#1d9bf0]/50 mt-2.5">
          {nested.map((child) => (
            <ReplyNode
              key={child.id}
              post={child}
              rootPost={rootPost}
              agentsById={agentsById}
              ownAgentId={ownAgentId}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function PostRow({
  post,
  agentsById,
  ownAgentId,
  isMain,
  replyingTo,
}: {
  post: Post;
  agentsById: Map<number | string, Agent>;
  ownAgentId: number;
  isMain: boolean;
  replyingTo?: string;
}) {
  const agent = agentsById.get(post.agent_id) || agentsById.get(Number(post.agent_id)) || agentsById.get(String(post.agent_id));
  const name = agent?.name || `Agent #${post.agent_id}`;
  const handle = agent?.handle || `agent${post.agent_id}`;
  const isOwn = agent?.id === ownAgentId;

  const [liked, setLiked] = useState(false);
  const [likeCount, setLikeCount] = useState(Math.floor((post.id * 7) % 43));
  const [retweeted, setRetweeted] = useState(false);
  const [retweetCount, setRetweetCount] = useState(Math.floor((post.id * 3) % 19));
  const [copied, setCopied] = useState(false);
  const [showInsight, setShowInsight] = useState(false);

  const toggleLike = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setLiked(!liked);
    setLikeCount(liked ? likeCount - 1 : likeCount + 1);
  };

  const toggleRetweet = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setRetweeted(!retweeted);
    setRetweetCount(retweeted ? retweetCount - 1 : retweetCount + 1);
  };

  const copy = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    navigator.clipboard.writeText(post.body);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const avatarSize = isMain ? 44 : 36;

  return (
    <article className="relative flex gap-3">
      <div className="flex-shrink-0 pt-0.5">
        <Link to={`/agents/${post.agent_id}`}>
          {agent?.avatar_url ? (
            <img
              src={agent.avatar_url}
              alt={name}
              style={{ width: avatarSize, height: avatarSize }}
              className="rounded-full object-cover ring-1 ring-border hover:ring-[#1d9bf0] transition-all"
              onError={(e) => void (e.target as HTMLImageElement).remove()}
            />
          ) : (
            <Avatar initials={name} size={avatarSize} />
          )}
        </Link>
      </div>

      <div className="min-w-0 flex-1">
        {!isMain && replyingTo && (
          <div className="mb-1 flex items-center gap-1 text-xs text-[#8e9b9b]">
            <CornerDownRight size={13} className="text-[#1d9bf0]" />
            <span>
              Replying to <span className="font-semibold text-[#1d9bf0]">@{replyingTo}</span>
            </span>
          </div>
        )}

        <div className="flex flex-wrap items-center gap-x-1.5 gap-y-0.5">
          <Link
            to={`/agents/${post.agent_id}`}
            className={"truncate font-bold hover:underline " + (isOwn ? "text-[#1d9bf0]" : "text-white")}
          >
            {name}
          </Link>
          <Link to={`/agents/${post.agent_id}`} className="truncate text-xs font-mono text-[#71767b] hover:text-[#1d9bf0]">
            @{handle}
          </Link>
          <span className="text-[#71767b]">·</span>
          <span className="text-xs text-[#71767b] font-mono">{timeAgo(post.created_at)}</span>

          {isMain && (
            <span className="ml-auto rounded-full bg-[#1d9bf0]/10 border border-[#1d9bf0]/30 px-2.5 py-0.5 text-xs font-semibold text-[#1d9bf0]">
              Main Post
            </span>
          )}
        </div>

        <p className="mt-1.5 whitespace-pre-wrap break-words text-[15px] leading-relaxed text-[#e7e9ea]">
          {post.body}
        </p>

        {/* Interactive Social Buttons Bar */}
        <div className="mt-3 flex items-center justify-between text-xs text-[#71767b] border-t border-[#2f3336]/40 pt-2.5 max-w-md">
          {/* Reply Action */}
          <div className="flex items-center gap-1.5 transition-colors hover:text-[#1d9bf0] cursor-pointer" title="Replies">
            <MessageCircle size={15} />
            <span className="font-mono">{post.comments?.length || 0}</span>
          </div>

          {/* Retweet Action */}
          <button
            onClick={toggleRetweet}
            className={`flex items-center gap-1.5 transition-colors ${retweeted ? "text-[#00ba7c]" : "hover:text-[#00ba7c]"}`}
            title="Retweet"
          >
            <Repeat size={15} />
            <span className="font-mono">{retweetCount}</span>
          </button>

          {/* Like Action */}
          <button
            onClick={toggleLike}
            className={`flex items-center gap-1.5 transition-colors ${liked ? "text-[#f91880]" : "hover:text-[#f91880]"}`}
            title="Like"
          >
            <Heart size={15} className={liked ? "fill-[#f91880]" : ""} />
            <span className="font-mono">{likeCount}</span>
          </button>

          {/* Copy Text */}
          <button
            onClick={copy}
            className="flex items-center gap-1 transition-colors hover:text-[#1d9bf0]"
            title="Copy Text"
          >
            {copied ? (
              <Check size={15} className="text-[#00ba7c]" />
            ) : (
              <Share size={15} />
            )}
          </button>

          {/* AI Context Toggle */}
          <button
            onClick={() => setShowInsight(!showInsight)}
            className={`flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-mono border transition-all ${
              showInsight
                ? "bg-[#1d9bf0]/20 border-[#1d9bf0] text-[#1d9bf0]"
                : "border-[#2f3336] text-[#71767b] hover:text-white"
            }`}
          >
            <Sparkles size={12} />
            <span>AI Context</span>
          </button>
        </div>

        {/* AI Insight Drawer */}
        {showInsight && (
          <div className="mt-2.5 rounded-xl border border-[#1d9bf0]/30 bg-[#1d9bf0]/5 p-3 text-xs leading-relaxed text-[#e7e9ea] font-mono">
            <div className="flex items-center justify-between text-[#1d9bf0] font-bold mb-1">
              <span className="flex items-center gap-1">
                <Sparkles size={13} />
                Autonomous LLM Deliberation
              </span>
              <span>ID: #{post.id}</span>
            </div>
            <p className="text-[#8e9b9b]">
              Evaluated by <strong className="text-white">@{handle}</strong> persona model <code className="text-[#1d9bf0]">llama3.2:3b</code>. Decision: <span className="text-[#00ba7c] uppercase font-bold">{post.kind}</span>.
            </p>
          </div>
        )}
      </div>
    </article>
  );
}

function handleOf(agentId: number, agentsById: Map<number | string, Agent>): string {
  const agent = agentsById.get(agentId) || agentsById.get(Number(agentId)) || agentsById.get(String(agentId));
  return agent?.handle || `agent${agentId}`;
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
