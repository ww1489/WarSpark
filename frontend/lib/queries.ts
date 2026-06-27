import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getBattleLog,
  getCWLGroup,
  getCapitalRaidSeasons,
  getClan,
  getClanBuilderBaseRanking,
  getClanCapitalRanking,
  getClanRanking,
  getCurrentGoldPass,
  getCurrentWar,
  getLeagues,
  getLocations,
  getPlayer,
  getPlayerBuilderBaseRanking,
  getPlayerLeagueGroup,
  getPlayerRanking,
  getWarLog,
  searchClans,
} from "./api";
import type { ClanSearchParams } from "./types";
import { useState, useCallback } from "react";

const STALE = {
  clan: 5 * 60 * 1000,
  player: 5 * 60 * 1000,
  warLog: 5 * 60 * 1000,
  ranking: 10 * 60 * 1000,
  goldPass: 60 * 60 * 1000,
  search: 2 * 60 * 1000,
};

export function useClan(tag: string) {
  return useQuery({
    queryKey: ["clan", tag],
    queryFn: () => getClan(tag),
    staleTime: STALE.clan,
    enabled: !!tag,
  });
}

export function usePlayer(tag: string) {
  return useQuery({
    queryKey: ["player", tag],
    queryFn: () => getPlayer(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  });
}

export function useWarLog(tag: string, limit = 10) {
  return useQuery({
    queryKey: ["warLog", tag, limit],
    queryFn: () => getWarLog(tag, limit),
    staleTime: STALE.warLog,
    enabled: !!tag,
  });
}

export function useRaidSeasons(tag: string) {
  return useQuery({
    queryKey: ["raidSeasons", tag],
    queryFn: () => getCapitalRaidSeasons(tag),
    staleTime: STALE.clan,
    enabled: !!tag,
  });
}

export function useBattleLog(tag: string) {
  return useQuery({
    queryKey: ["battleLog", tag],
    queryFn: () => getBattleLog(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  });
}

export function useLeagueGroup(tag: string) {
  return useQuery({
    queryKey: ["leagueGroup", tag],
    queryFn: () => getPlayerLeagueGroup(tag),
    staleTime: STALE.player,
    enabled: !!tag,
  });
}

export function useLocations() {
  return useQuery({
    queryKey: ["locations"],
    queryFn: () => getLocations(),
    staleTime: STALE.ranking,
  });
}

export type RankingType = "clans" | "players" | "capital" | "builder-clans" | "builder-players";

export function useRankings(locationId: string, type: RankingType) {
  return useQuery({
    queryKey: ["rankings", locationId, type],
    queryFn: () => {
      switch (type) {
        case "clans":
          return getClanRanking(locationId, 100);
        case "players":
          return getPlayerRanking(locationId, 100);
        case "capital":
          return getClanCapitalRanking(locationId, 100);
        case "builder-clans":
          return getClanBuilderBaseRanking(locationId, 100);
        case "builder-players":
          return getPlayerBuilderBaseRanking(locationId, 100);
      }
    },
    staleTime: STALE.ranking,
    enabled: !!locationId,
  });
}

export const HOMEPAGE_LIMIT = 10;

export function useHomeClanRankings(locationId: string) {
  return useQuery({
    queryKey: ["rankings", locationId, "clans", "home"],
    queryFn: () => getClanRanking(locationId, HOMEPAGE_LIMIT),
    staleTime: STALE.ranking,
    enabled: !!locationId,
  });
}

export function useHomePlayerRankings(locationId: string) {
  return useQuery({
    queryKey: ["rankings", locationId, "players", "home"],
    queryFn: () => getPlayerRanking(locationId, HOMEPAGE_LIMIT),
    staleTime: STALE.ranking,
    enabled: !!locationId,
  });
}

export function useHomeCapitalRankings(locationId: string) {
  return useQuery({
    queryKey: ["rankings", locationId, "capital", "home"],
    queryFn: () => getClanCapitalRanking(locationId, HOMEPAGE_LIMIT),
    staleTime: STALE.ranking,
    enabled: !!locationId,
  });
}

export function useSearchClans(params: ClanSearchParams) {
  return useQuery({
    queryKey: ["searchClans", params],
    queryFn: () => searchClans(params),
    staleTime: STALE.search,
    enabled: !!(params.name || params.locationId || params.warFrequency),
  });
}

export function useCurrentWar(tag: string) {
  return useQuery({
    queryKey: ["currentWar", tag],
    queryFn: () => getCurrentWar(tag),
    refetchInterval: 30_000,
    staleTime: 30_000,
    enabled: !!tag,
  });
}

export function useCWLGroup(tag: string) {
  return useQuery({
    queryKey: ["cwlGroup", tag],
    queryFn: () => getCWLGroup(tag),
    staleTime: 5 * 60 * 1000,
    enabled: !!tag,
  });
}

export function useCurrentGoldPass() {
  return useQuery({
    queryKey: ["goldPass"],
    queryFn: () => getCurrentGoldPass(),
    staleTime: STALE.goldPass,
  });
}

export function useAllLeagues() {
  return useQuery({
    queryKey: ["leagues"],
    queryFn: () => getLeagues(),
    staleTime: 24 * 60 * 60 * 1000,
  });
}
