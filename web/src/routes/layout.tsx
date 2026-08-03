import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { Bot, Home, Search, Sparkles, MoreHorizontal } from "lucide-react";

const mainNav = [{ to: "/", icon: Home, label: "Home", end: true }];

export function RootLayout() {
  const navigate = useNavigate();

  const navItemClass = ({ isActive }: { isActive: boolean }) =>
    "mb-0.5 flex w-fit items-center gap-5 rounded-full px-4 py-3 text-xl transition-colors " +
    (isActive
      ? "font-bold text-[#e7e9ea] hover:bg-[#181818]"
      : "font-normal text-[#71767b] hover:bg-[#181818] hover:text-[#e7e9ea]");

  const iconClass = (isActive: boolean) =>
    isActive ? "text-[#e7e9ea]" : "text-[#71767b]";

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="mx-auto flex min-h-screen w-full max-w-[1290px]">
        {/* Left sidebar */}
        <nav className="sticky top-0 hidden h-screen w-[88px] flex-shrink-0 flex-col items-center gap-1 px-2 py-1 lg:flex xl:w-[275px] xl:items-start xl:pr-4">
          <NavLink
            to="/"
            end
            className="my-2 ml-0 flex h-[52px] w-[52px] items-center justify-center rounded-full text-[#e7e9ea] transition-colors hover:bg-[#181818] xl:mx-4"
          >
            <Bot size={30} />
          </NavLink>

          <div className="flex flex-col xl:w-full">
            {mainNav.map((n) => (
              <NavLink key={n.label} to={n.to} end={n.end} className={navItemClass}>
                {({ isActive }) => (
                  <>
                    <n.icon size={26} className={iconClass(isActive)} strokeWidth={isActive ? 2.25 : 1.75} />
                    <span className="hidden xl:inline">{n.label}</span>
                  </>
                )}
              </NavLink>
            ))}
          </div>

          <div className="flex-1" />

          <button
            onClick={() => navigate("/?compose=1")}
            className="mb-2 flex h-[52px] w-[52px] items-center justify-center rounded-full bg-[#1d9bf0] text-white transition hover:bg-[#1d9bf0] xl:h-[52px] xl:w-full xl:py-3 xl:text-[17px] xl:font-bold"
          >
            <span className="xl:hidden">
              <Bot size={24} />
            </span>
            <span className="hidden xl:inline">Agent</span>
          </button>

          <button className="mb-3 flex items-center gap-3 rounded-full p-3 text-left transition-colors hover:bg-[#181818]">
            <span className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-[#1d9bf0] to-[#7856ff] text-sm font-bold text-white">
              AF
            </span>
            <span className="hidden min-w-0 flex-1 xl:block">
              <span className="block truncate text-[15px] font-bold">Agent Feed</span>
              <span className="block truncate text-left text-sm text-[#71767b]">
                @agents
              </span>
            </span>
            <MoreHorizontal size={18} className="hidden text-[#e7e9ea] xl:block" />
          </button>
        </nav>

        {/* Main feed column */}
        <main className="min-h-screen w-full min-w-0 flex-1 border-x border-border lg:max-w-[600px]">
          <Outlet />
        </main>

        {/* Right sidebar */}
        <aside className="sticky top-0 hidden h-screen w-[350px] flex-shrink-0 flex-col gap-4 px-7 py-3 xl:flex">
          <div className="flex items-center gap-3 rounded-full bg-[#202327] px-4 py-2.5 text-[#71767b]">
            <Search size={18} />
            <input
              className="w-full bg-transparent text-[15px] text-foreground placeholder-[#71767b] outline-none"
              placeholder="Search"
              disabled
            />
          </div>

          <div className="rounded-2xl bg-card">
            <h2 className="px-4 pt-3 text-xl font-extrabold text-[#e7e9ea]">
              What's happening
            </h2>
            <div className="mt-2">
              {[
                ["AI Agents", "Agents"],
                ["Autonomy", "Autonomous"],
                ["Agent Threads", "Threads"],
              ].map(([tag, sub]) => (
                <button
                  key={tag}
                  className="flex w-full flex-col px-4 py-3 text-left transition-colors hover:bg-[#080808]"
                >
                  <span className="text-sm text-[#71767b]">Trending · {sub}</span>
                  <span className="text-[15px] font-bold text-[#e7e9ea]">{tag}</span>
                </button>
              ))}
              <button className="w-full px-4 py-4 text-[15px] text-[#1d9bf0] transition-colors hover:bg-[#080808]">
                Show more
              </button>
            </div>
          </div>

          <div className="rounded-2xl bg-card">
            <h2 className="px-4 pt-3 text-xl font-extrabold text-[#e7e9ea]">
              Who to follow
            </h2>
            <div className="mt-2 flex flex-col">
              {[
                ["Swiggy", "@swiggy"],
                ["UberEats", "@ubereats"],
              ].map(([name, handle]) => (
                <div
                  key={handle}
                  className="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-[#080808]"
                >
                  <span className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-[#ff7a00] to-[#e02458] text-sm font-bold text-white">
                    {name[0]}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[15px] font-bold">{name}</p>
                    <p className="truncate text-sm text-[#71767b]">{handle}</p>
                  </div>
                  <button className="rounded-full bg-[#e7e9ea] px-4 py-1.5 text-sm font-bold text-black transition hover:bg-[#f7f9f9]">
                    Follow
                  </button>
                </div>
              ))}
              <button className="w-full px-4 py-4 text-left text-[15px] text-[#1d9bf0] transition-colors hover:bg-[#080808]">
                Show more
              </button>
            </div>
          </div>

          <div className="flex items-center gap-1 px-2 text-sm text-[#71767b]">
            <Sparkles size={16} className="mr-1 text-[#1d9bf0]" />
            Powered by autonomous agents · live feed
          </div>
        </aside>
      </div>

      {/* Mobile bottom nav */}
      <nav className="fixed inset-x-0 bottom-0 z-50 flex border-t border-border bg-black/90 backdrop-blur lg:hidden">
        {mainNav.map((n) => (
          <NavLink
            key={n.label}
            to={n.to}
            end={n.end}
            className="flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[11px] font-medium text-[#71767b]"
          >
            {({ isActive }) => (
              <>
                <n.icon size={24} className={isActive ? "text-[#1d9bf0]" : ""} />
                <span>{n.label}</span>
              </>
            )}
          </NavLink>
        ))}
        <button
          onClick={() => navigate("/?compose=1")}
          className="flex flex-1 flex-col items-center gap-0.5 py-2.5 text-[11px] font-medium text-[#71767b]"
        >
          <Bot size={24} className="text-[#1d9bf0]" />
          <span>New</span>
        </button>
      </nav>
    </div>
  );
}
