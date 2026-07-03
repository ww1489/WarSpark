// === Common ===
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface Paging {
  cursors: { after?: string; before?: string };
}

export interface PagedList<T> {
  items: T[];
  paging: Paging;
}

export interface Label {
  id: number;
  name: string;
  iconUrls: { small: string; medium: string };
}

export interface Location {
  id: number;
  name: string;
  countryCode?: string;
}

export interface LeagueRef {
  id: number;
  name: string;
}

// === Clan ===
export interface ClanOverview {
  tag: string;
  name: string;
  clanLevel: number;
  description: string;
  members: number;
  clanPoints: number;
  warWins: number;
  warLosses?: number;
  warTies?: number;
  warWinStreak: number;
  warFrequency: string;
  type: string;
  isWarLogPublic: boolean;
  location?: Location;
  warLeague?: LeagueRef;
  badgeUrls: { small: string; medium: string; large: string };
  labels: Label[];
}

export interface ClanDetail {
  clan: ClanOverview;
  members: ClanMemberSummary[];
}

export interface ClanMemberSummary {
  tag: string;
  name: string;
  role: string;
  expLevel: number;
  townHallLevel: number;
  trophies: number;
  clanRank: number;
  donations: number;
  donationsReceived: number;
  league?: LeagueRef;
  leagueTier?: { id: number; name: string; iconUrls?: { small: string; large: string } };
}

// === Player ===
export interface PlayerOverview {
  tag: string;
  name: string;
  townHallLevel: number;
  townHallWeaponLevel?: number;
  expLevel: number;
  role?: string;
  warStars?: number;
  attackWins?: number;
  defenseWins?: number;
  trophies: number;
  bestTrophies?: number;
  warPreference?: string;
  builderBaseTrophies?: number;
  bestBuilderBaseTrophies?: number;
  builderHallLevel?: number;
  donations?: number;
  donationsReceived?: number;
  clanCapitalContributions?: number;
  clan?: PlayerClanInfo;
  league?: LeagueRef;
  leagueTier?: { id: number; name: string; iconUrls?: { small: string; large: string } };
  builderBaseLeague?: LeagueRef;
  labels?: Label[];
  heroes?: HeroLevel[];
  heroEquipment?: HeroLevel[];
  achievements?: AchievementProgress[];
  troops?: TroopSpellLevel[];
  spells?: TroopSpellLevel[];
  legendStatistics?: PlayerLegendStatistics;
}

export interface PlayerClanInfo {
  tag: string;
  name: string;
  clanLevel: number;
  badgeUrls: { small: string; medium: string };
}

export interface HeroLevel {
  name: string;
  level: number;
  maxLevel: number;
  village: string;
}

export interface TroopSpellLevel {
  name: string;
  level: number;
  maxLevel: number;
  village?: string;
  superTroopActive?: boolean;
  equipment?: HeroLevel[];
}

export interface PlayerLegendStatistics {
  legendTrophies?: number;
  bestSeason?: LegendSeasonResult;
  currentSeason?: LegendSeasonResult;
  previousSeason?: LegendSeasonResult;
}

export interface LegendSeasonResult {
  rank?: number;
  trophies?: number;
}

export interface AchievementProgress {
  name: string;
  stars: number;
  target: number;
  value: number;
  village: string;
  info: string;
}

// === War ===
export interface WarLogEntry {
  result: string;
  teamSize: number;
  endTime: string;
  clan: WarLogClan;
  opponent: WarLogClan;
}

export interface WarLogClan {
  tag: string;
  name: string;
  clanLevel: number;
  stars: number;
  destructionPercentage: number;
  attacks: number;
}

export interface CurrentWar {
  state: string;
  teamSize: number;
  clan: WarClan;
  opponent: WarClan;
}

export interface WarClan {
  tag: string;
  name: string;
  stars: number;
  destructionPercentage: number;
  clanLevel: number;
  attacks: number;
}

export interface CWLGroup {
  state: string;
  season: string;
  clans: CWLClan[];
  rounds: CWLRound[];
}

export interface CWLClan {
  tag: string;
  name: string;
  clanLevel: number;
  members: number;
}

export interface CWLRound {
  warTags: string[];
}

// === Rankings ===
export interface ClanRankingEntry {
  tag: string;
  name: string;
  location: Location;
  clanLevel: number;
  members: number;
  clanPoints: number;
  rank: number;
  previousRank: number;
  badgeUrls: { small: string; medium: string };
}

export interface PlayerRankingEntry {
  tag: string;
  name: string;
  expLevel: number;
  trophies: number;
  rank: number;
  previousRank: number;
  attackWins: number;
  defenseWins: number;
  clan?: { tag: string; name: string; badgeUrls: { small: string; medium: string } };
  leagueTier?: { id: number; name: string; iconUrls?: { small: string; large: string } };
}

export interface ClanCapitalEntry {
  tag: string;
  name: string;
  location: Location;
  clanLevel: number;
  members: number;
  clanCapitalPoints: number;
  rank: number;
  previousRank: number;
  badgeUrls: { small: string; medium: string };
}

export interface ClanBuilderBaseEntry {
  tag: string;
  name: string;
  location: Location;
  clanLevel: number;
  members: number;
  clanBuilderBasePoints: number;
  rank: number;
  previousRank: number;
  badgeUrls: { small: string; medium: string };
}

export interface PlayerBuilderBaseEntry {
  tag: string;
  name: string;
  expLevel: number;
  builderBaseTrophies: number;
  rank: number;
  previousRank: number;
  clan?: { tag: string; name: string; badgeUrls: { small: string; medium: string } };
}

// === Battle Log ===
export interface BattleLogEntry {
  battleTime: number | string;
  battleType: string;
  attack: boolean;
  stars: number;
  destructionPercentage: number;
  opponentName: string;
  opponentTag: string;
  opponentTownHallLevel: number;
}

export interface BattleLogSummary {
  items: BattleLogEntry[];
}

// === Gold Pass ===
export interface GoldPassSeason {
  startTime: string;
  endTime: string;
}

// === Capital Raids ===
export interface CapitalRaidSeason {
  state: string;
  startTime: string;
  endTime: string;
  capitalTotalLoot: number;
  raidsCompleted: number;
  totalAttacks: number;
  offensiveReward: number;
  defensiveReward: number;
  enemyDistrictsDestroyed: number;
}

// === League Group ===
export interface PlayerLeagueGroup {
  members: LeagueGroupMember[];
  attackLogs: LeagueBattleLogEntry[];
  defenseLogs: LeagueBattleLogEntry[];
}

export interface LeagueGroupMember {
  playerName: string;
  playerTag: string;
  clanName: string;
  clanTag: string;
  leagueTrophies: number;
  attackWinCount: number;
  attackLoseCount: number;
  defenseWinCount: number;
  defenseLoseCount: number;
}

export interface LeagueBattleLogEntry {
  creationTime: string;
  destructionPercentage: number;
  opponentName: string;
  stars: number;
  trophies: number;
}

// === Search ===
export interface LeagueInfo {
  id: number;
  name: string;
  iconUrls?: { small: string; medium: string; tiny: string };
}

export interface ClanSearchParams {
  name?: string;
  warFrequency?: string;
  locationId?: number;
  minMembers?: number;
  maxMembers?: number;
  minClanPoints?: number;
  minClanLevel?: number;
  limit?: number;
  after?: string;
  before?: string;
  labelIds?: string;
}

// === Image Search / Find Layout ===
export interface ImageSearchJob {
  job_id: string;
  search_status: string;
  detected_th?: number;
  screenshot_quality?: string;
  buildings_detected?: number;
  error_code?: string;
  error_message?: string;
  created_at: string;
  updated_at?: string;
}

export interface ImageSearchResults {
  job: ImageSearchJob;
  match_summary: { candidate_count: number; best_match_level?: string; low_confidence: boolean };
  layouts: LayoutCard[];
  attack_videos: VideoMatch[];
  defense_replays: VideoMatch[];
}

export interface LayoutCard {
  layout_id: string;
  title: string;
  th_level: number;
  layout_type: string;
  style_tags?: string[];
  primary_image_url?: string;
  source_type?: string;
}

export interface LayoutDetail {
  layout_id: string;
  title: string;
  th_level: number;
  layout_type: string;
  style_tags?: string[];
  source_type?: string;
  source_url?: string;
  images: LayoutImage[];
  links: LayoutLink[];
  attack_videos: VideoMatch[];
  defense_replays: VideoMatch[];
  similar_layouts: LayoutCard[];
}

export interface LayoutImage {
  image_id: string;
  image_url: string;
  width?: number;
  height?: number;
  image_role: string;
}

export interface LayoutLink {
  link_id: string;
  link_type: string;
  url: string;
  link_status: string;
}

export interface VideoMatch {
  match_id: string;
  video_id: string;
  youtube_video_id: string;
  video_title: string;
  channel_name?: string;
  timestamp_seconds: number;
  youtube_url: string;
  match_group: string;
  match_type: string;
}
