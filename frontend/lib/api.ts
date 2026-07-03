import type { ApiResponse, GoldPassSeason, PagedList } from "./types";
import type {
  BattleLogSummary,
  CWLGroup,
  CapitalRaidSeason,
  ClanBuilderBaseEntry,
  ClanCapitalEntry,
  ClanDetail,
  ClanOverview,
  ClanRankingEntry,
  ClanSearchParams,
  CurrentWar,
  ImageSearchJob,
  ImageSearchResults,
  LayoutCard,
  LayoutDetail,
  LeagueInfo,
  Location,
  PlayerBuilderBaseEntry,
  PlayerLeagueGroup,
  PlayerOverview,
  PlayerRankingEntry,
  VideoMatch,
  WarLogEntry,
} from "./types";
import { ApiError } from "./types";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

async function fetchApi<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) throw new ApiError(res.status, await res.text());
  const json: ApiResponse<T> = await res.json();
  if (json.code !== 0) throw new ApiError(json.code, json.message);
  return json.data;
}

function enc(tag: string) {
  return encodeURIComponent(tag);
}

// Clan
export async function getClan(tag: string) {
  return fetchApi<ClanDetail>(`/clans/${enc(tag)}`);
}

export async function getWarLog(tag: string, limit = 10, after?: string) {
  const params = new URLSearchParams({ limit: String(limit) });
  if (after) params.set("after", after);
  return fetchApi<PagedList<WarLogEntry>>(`/clans/${enc(tag)}/war-log?${params}`);
}

export async function getCapitalRaidSeasons(tag: string) {
  return fetchApi<PagedList<CapitalRaidSeason>>(`/clans/${enc(tag)}/capital-raid-seasons`);
}

// Player
export async function getPlayer(tag: string) {
  return fetchApi<PlayerOverview>(`/players/${enc(tag)}`);
}

export async function getBattleLog(tag: string) {
  return fetchApi<BattleLogSummary>(`/players/${enc(tag)}/battle-log`);
}

export async function getPlayerLeagueGroup(tag: string) {
  return fetchApi<PlayerLeagueGroup>(`/players/${enc(tag)}/league-group`);
}

// Rankings
export async function getLocations() {
  return fetchApi<PagedList<Location>>("/locations");
}

export async function getClanRanking(locationId: string, limit = 200) {
  return fetchApi<PagedList<ClanRankingEntry>>(`/locations/${locationId}/rankings/clans?limit=${limit}`);
}

export async function getPlayerRanking(locationId: string, limit = 200) {
  return fetchApi<PagedList<PlayerRankingEntry>>(`/locations/${locationId}/rankings/players?limit=${limit}`);
}

export async function getClanCapitalRanking(locationId: string, limit = 200) {
  return fetchApi<PagedList<ClanCapitalEntry>>(`/locations/${locationId}/rankings/clans-capital?limit=${limit}`);
}

export async function getClanBuilderBaseRanking(locationId: string, limit = 200) {
  return fetchApi<PagedList<ClanBuilderBaseEntry>>(`/locations/${locationId}/rankings/clans-builder-base?limit=${limit}`);
}

export async function getPlayerBuilderBaseRanking(locationId: string, limit = 200) {
  return fetchApi<PagedList<PlayerBuilderBaseEntry>>(`/locations/${locationId}/rankings/players-builder-base?limit=${limit}`);
}

// Leagues
export async function getLeagues() {
  return fetchApi<PagedList<LeagueInfo>>("/leagues?limit=100");
}

// Search
export async function searchClans(params: ClanSearchParams) {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v != null && v !== "") sp.set(k, String(v));
  }
  return fetchApi<PagedList<ClanOverview>>(`/clans?${sp}`);
}

// War
export async function getCurrentWar(clanTag: string) {
  return fetchApi<CurrentWar>(`/war/current?clan_tag=${enc(clanTag)}`);
}

export async function getCWLGroup(clanTag: string) {
  return fetchApi<CWLGroup>(`/war/cwl?clan_tag=${enc(clanTag)}`);
}

export async function getCWLWar(warTag: string) {
  return fetchApi<CurrentWar>(`/war/cwl/wars/${enc(warTag)}`);
}

// Gold Pass
export async function getCurrentGoldPass() {
  return fetchApi<GoldPassSeason>("/gold-pass/current");
}

// Image Search (找阵)
export async function createImageSearchJob(file: File, thLevel?: number) {
  const form = new FormData();
  form.append("image", file);
  if (thLevel) form.append("th_level", String(thLevel));
  return fetchApi<ImageSearchJob>("/image-search/jobs", { method: "POST", body: form, headers: {} });
}

export async function getImageSearchJob(jobId: string) {
  return fetchApi<ImageSearchJob>(`/image-search/jobs/${jobId}`);
}

export async function getImageSearchResults(jobId: string) {
  return fetchApi<ImageSearchResults>(`/image-search/jobs/${jobId}/results`);
}

// Layouts (阵型)
export async function listLayouts(params?: {
  th_level?: number;
  layout_type?: string;
  page?: number;
  per_page?: number;
}) {
  const sp = new URLSearchParams();
  if (params?.th_level) sp.set("th_level", String(params.th_level));
  if (params?.layout_type) sp.set("layout_type", params.layout_type);
  if (params?.page) sp.set("page", String(params.page));
  if (params?.per_page) sp.set("per_page", String(params.per_page));
  const qs = sp.toString();
  return fetchApi<{ items: LayoutCard[]; total: number }>(`/layouts${qs ? `?${qs}` : ""}`);
}

export async function getLayout(layoutId: string) {
  return fetchApi<LayoutDetail>(`/layouts/${layoutId}`);
}

export async function getLayoutVideos(layoutId: string, matchGroup?: string, matchType?: string) {
  const sp = new URLSearchParams();
  if (matchGroup) sp.set("match_group", matchGroup);
  if (matchType) sp.set("match_type", matchType);
  const qs = sp.toString();
  return fetchApi<VideoMatch[]>(`/layouts/${layoutId}/videos${qs ? `?${qs}` : ""}`);
}
