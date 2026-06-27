import { Outlet } from "@tanstack/react-router";

export default function AdminLayout() {
  return (
    <div>
      <div className="bg-surface-card border-b border-border">
        <div className="max-w-page mx-auto flex gap-6 px-4 h-12 items-center text-sm">
          <a href="/admin/war" className="text-text-secondary hover:text-primary font-medium">战争监控</a>
          <a href="/admin/cwl" className="text-text-secondary hover:text-primary font-medium">CWL 联赛</a>
        </div>
      </div>
      <Outlet />
    </div>
  );
}
