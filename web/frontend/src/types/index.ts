export interface Game {
  gameCode: string;
  leagueId: number;
  leagueName: string;
  currentState: GameState;
  isFetching: boolean;
  extTimestamp: string;
  // Enhanced fields
  gameDetails?: GameDetails;
  statistics?: GameStats;
  events?: GameEvent[];
}

export interface GameState {
  home: TeamState;
  away: TeamState;
  status: GameStatus;
  fetchedAt: string;
  extTimestamp?: string;
  // Enhanced game state
  period?: number;
  periodType?: string;
  timeRemaining?: string;
  clock?: string;
  venue?: Venue;
  weather?: Weather;
  // Baseball-specific details
  details?: EventDetails;
  statistics?: TeamStats;
}

export interface TeamState {
  team: Team;
  score: number;
  // Enhanced team state
  periodScores?: number[];
  statistics?: TeamStats;
}

export interface Team {
  teamId: number;
  teamCode: string;
  teamName: string;
  leagueId: number;
  extId: string;
  logoUrl?: string;
}

export interface GameDetails {
  gameId: string;
  season: string;
  seasonType: string;
  week?: number;
  gameDate: string;
  gameTime: string;
  timezone: string;
  broadcasters?: Broadcaster[];
  officials?: Official[];
}

export interface Venue {
  id: string;
  name: string;
  city: string;
  state: string;
  country: string;
  capacity?: number;
  surface?: string;
  indoor?: boolean;
}

export interface Weather {
  temperature?: number;
  condition?: string;
  humidity?: number;
  windSpeed?: number;
  windDirection?: string;
}

export interface Broadcaster {
  name: string;
  network: string;
  language: string;
}

export interface Official {
  name: string;
  position: string;
  number?: number;
}

export interface GameStats {
  totalPlays: number;
  totalYards?: number;
  timeOfPossession?: string;
  turnovers: number;
  penalties: number;
  penaltyYards?: number;
}

export interface TeamStats {
  // General stats
  plays: number;
  yards?: number;
  timeOfPossession?: string;
  turnovers: number;
  penalties: number;
  penaltyYards?: number;
  
  // League-specific stats
  firstDowns?: number;
  rushingYards?: number;
  passingYards?: number;
  shots?: number;
  hits?: number;
  faceoffs?: number;
  powerPlays?: number;
  errors?: number;
  strikeouts?: number;
  walks?: number;
}

export interface GameEvent {
  id: string;
  type: EventType;
  period: number;
  time: string;
  clock?: string;
  description: string;
  team: Team;
  player?: Player;
  details?: EventDetails;
  timestamp: string;
}

export type EventType = 
  | "goal"
  | "assist"
  | "penalty"
  | "power_play"
  | "shot"
  | "hit"
  | "faceoff"
  | "save"
  | "turnover"
  | "fumble"
  | "interception"
  | "touchdown"
  | "field_goal"
  | "safety"
  | "home_run"
  | "strikeout"
  | "walk"
  | "error"
  | "period_start"
  | "period_end"
  | "game_start"
  | "game_end";

export interface Player {
  id: string;
  name: string;
  number: number;
  position: string;
  team: Team;
}

export interface BaseRunners {
  first?: Player;
  second?: Player;
  third?: Player;
}

export interface EventDetails {
  // Goal details
  goalType?: string;
  assist1?: Player;
  assist2?: Player;
  
  // Penalty details
  penaltyType?: string;
  penaltyMinutes?: number;
  
  // Play details
  yardLine?: number;
  down?: number;
  distance?: number;
  yardsGained?: number;
  possession?: string; // Team code with possession
  
  // Baseball details
  inning?: number;
  outs?: number;
  bases?: string;
  pitchCount?: number;
  strikeCount?: number;
  ballCount?: number;
  // Enhanced baseball details
  baseRunners?: BaseRunners;
  pitcher?: Player;
  batter?: Player;
}

export enum GameStatus {
  Upcoming = "upcoming",
  Active = "active", 
  Delayed = "delayed",
  Ended = "ended"
}

export interface Event {
  // Basic event info
  id: string;
  type: EventType;
  timestamp: string;
  description: string;
  
  // Team and player info
  teamCode: string;
  teamName: string;
  teamHash: string;
  playerName?: string;
  playerNumber?: number;
  
  // Game context
  leagueId: number;
  leagueName: string;
  gameCode: string;
  gameId: string;
  period: number;
  time: string;
  clock?: string;
  
  // Opponent info
  opponentCode: string;
  opponentName: string;
  opponentHash: string;
  
  // Event details
  details?: EventDetails;
  score?: ScoreUpdate;
  
  // Venue and broadcast info
  venue?: Venue;
  broadcasters?: Broadcaster[];
}

export interface ScoreUpdate {
  homeScore: number;
  awayScore: number;
  homeTeam: string;
  awayTeam: string;
}

export interface RichEvent extends Event {
  // Additional context for external systems
  gameState?: GameState;
  teamStats?: TeamStats;
  playerStats?: PlayerStats;
  weather?: Weather;
  broadcastInfo?: BroadcastInfo;
}

export interface PlayerStats {
  player: Player;
  goals?: number;
  assists?: number;
  points?: number;
  shots?: number;
  hits?: number;
  penaltyMinutes?: number;
  rushingYards?: number;
  passingYards?: number;
  receptions?: number;
  touchdowns?: number;
  rbis?: number;
  strikeouts?: number;
  walks?: number;
}

export interface BroadcastInfo {
  networks: string[];
  streaming?: string[];
  radio?: string[];
  language: string;
  availability: string;
}

export interface LeagueConfig {
  leagueId: number;
  leagueName: string;
  teams: string[];
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
}

export interface WebSocketMessage {
  type: 'game_update' | 'event' | 'games_list' | 'period_update' | 'game_start' | 'game_end';
  data: any;
}

// Event priority levels
export type EventPriority = 'low' | 'normal' | 'high' | 'critical';

// Utility functions for events
export const getEventIcon = (eventType: EventType): string => {
  switch (eventType) {
    case 'goal': return '🏒';
    case 'touchdown': return '🏈';
    case 'home_run': return '⚾';
    case 'penalty': return '⚠️';
    case 'power_play': return '⚡';
    case 'shot': return '🎯';
    case 'save': return '🛡️';
    case 'strikeout': return '⚡';
    case 'walk': return '🚶';
    case 'error': return '❌';
    case 'game_start': return '🏁';
    case 'game_end': return '🏁';
    case 'period_start': return '⏰';
    case 'period_end': return '⏰';
    default: return '📰';
  }
};

export const getEventColor = (eventType: EventType): string => {
  switch (eventType) {
    case 'goal':
    case 'touchdown':
    case 'home_run':
      return 'green';
    case 'penalty':
    case 'turnover':
    case 'fumble':
    case 'error':
      return 'red';
    case 'power_play':
    case 'strikeout':
      return 'yellow';
    case 'game_start':
    case 'game_end':
      return 'blue';
    case 'period_start':
    case 'period_end':
      return 'purple';
    default:
      return 'gray';
  }
};

export const getEventPriority = (eventType: EventType): EventPriority => {
  switch (eventType) {
    case 'goal':
    case 'touchdown':
    case 'home_run':
      return 'high';
    case 'game_start':
    case 'game_end':
    case 'period_start':
    case 'period_end':
      return 'normal';
    case 'penalty':
    case 'turnover':
    case 'fumble':
      return 'high';
    default:
      return 'normal';
  }
};