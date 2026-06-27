import { Outlet, useRouterState } from "@tanstack/react-router";

const navItems = [
  { to: "/", label: "首页" },
  { to: "/leaderboards", label: "排行榜" },
  { to: "/search", label: "部落搜索" },
  { to: "/gold-pass", label: "Gold Pass" },
];

export default function RootLayout() {
  const router = useRouterState();
  const currentPath = router.location.pathname;

  return (
    <div className="min-h-screen bg-surface-page flex flex-col">
      <nav className="bg-white/80 backdrop-blur-md border-b border-border sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center h-14 px-6">
          <a href="/" className="flex items-center gap-2 mr-10">
            <span className="text-xl font-bold text-primary tracking-tight">WarSpark</span>
          </a>
          <div className="flex gap-6">
            {navItems.map((item) => {
              const active = item.to === "/" ? currentPath === "/" : currentPath.startsWith(item.to);
              return (
                <a
                  key={item.to}
                  href={item.to}
                  className={`text-sm font-medium transition-colors ${
                    active ? "text-primary" : "text-text-secondary hover:text-text-primary"
                  }`}
                >
                  {item.label}
                </a>
              );
            })}
          </div>
          <div className="ml-auto" />
        </div>
      </nav>
      <main className="flex-1 pb-12">
        <Outlet />
      </main>
      <footer className="text-center text-text-muted text-xs py-5 border-t border-border-light bg-white">
        WarSpark · Clash of Clans 数据平台
      </footer>
    </div>
  );
}
