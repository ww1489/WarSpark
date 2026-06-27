export function normalizeTag(tag: string): string {
  let t = tag.trim().toUpperCase();
  if (!t.startsWith("#")) t = `#${t}`;
  return t;
}

export function formatNumber(n: number | undefined | null): string {
  if (n == null) return "0";
  return n.toLocaleString();
}

export function formatDate(iso: string | number): string {
  const d = parseCoCDate(iso);
  return `${d.getMonth() + 1}-${d.getDate()}`;
}

export function formatDateTime(iso: string | number): string {
  const d = parseCoCDate(iso);
  return `${d.getMonth() + 1}-${d.getDate()} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function parseCoCDate(raw: string | number): Date {
  if (typeof raw === "number") return new Date(raw > 1e12 ? raw : raw * 1000);
  const s = String(raw);
  const m = s.match(/^(\d{4})(\d{2})(\d{2})T(\d{2})(\d{2})(\d{2})/);
  if (m) return new Date(Date.UTC(+m[1], +m[2] - 1, +m[3], +m[4], +m[5], +m[6]));
  return new Date(s);
}

export function daysRemaining(endTime: string | number): number {
  const diff = parseCoCDate(endTime).getTime() - Date.now();
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)));
}

export function rankChangeText(rank: number, prev: number): string {
  if (prev <= 0) return "";
  const diff = prev - rank;
  if (diff > 0) return `+${diff}`;
  if (diff < 0) return `${diff}`;
  return "-";
}

export function warResultLabel(result: string): { text: string; color: string } {
  switch (result) {
    case "win":
      return { text: "胜利", color: "text-green-600" };
    case "lose":
      return { text: "失败", color: "text-red-600" };
    case "tie":
      return { text: "平局", color: "text-yellow-600" };
    default:
      return { text: result, color: "text-gray-500" };
  }
}

export function roleLabel(role: string): string {
  const m: Record<string, string> = {
    leader: "首领",
    coLeader: "副首领",
    admin: "长老",
    member: "成员",
  };
  return m[role] ?? role;
}

export function warStateLabel(state: string): string {
  const m: Record<string, string> = {
    notInWar: "未参战",
    preparation: "准备日",
    inWar: "战斗中",
    warEnded: "已结束",
  };
  return m[state] ?? state;
}

export function leagueIconUrl(league: { id?: number; name?: string }): string | null {
  if (!league?.id) return null;
  return `https://api-assets.clashofclans.com/leagues/${league.id}.png`;
}
