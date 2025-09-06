import React, { useState, useEffect } from 'react';
import { Game } from '../types';
import { apiClient } from '../utils/api';

interface ScoreboardProps {
  games: Game[];
}

interface UpcomingGame {
  gameCode: string;
  leagueId: number;
  leagueName: string;
  currentState: {
    home: { team: { teamCode: string; teamName: string; logoUrl?: string }; score: number };
    away: { team: { teamCode: string; teamName: string; logoUrl?: string }; score: number };
    status: string;
    clock?: string;
    venue?: { name: string; city?: string };
  };
}

const Scoreboard: React.FC<ScoreboardProps> = ({ games }) => {
  const [upcomingGames, setUpcomingGames] = useState<UpcomingGame[]>([]);
  const [showUpcoming, setShowUpcoming] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const gamesPerPage = 5;

  useEffect(() => {
    const fetchUpcomingGames = async () => {
      try {
        const response = await apiClient.get('/api/upcoming');
        if (response.data.success) {
          setUpcomingGames(response.data.data || []);
        }
      } catch (error) {
        console.error('Failed to fetch upcoming games:', error);
      }
    };

    fetchUpcomingGames();
  }, []);

  const getLeagueIcon = (leagueId: number) => {
    switch (leagueId) {
      case 1: return '🏒'; // NHL
      case 2: return '⚾'; // MLB
      case 5: return '🏈'; // CFL
      default: return '🏆';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500';
      case 'upcoming': return 'bg-blue-500';
      case 'ended': return 'bg-gray-500';
      default: return 'bg-gray-400';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return 'LIVE';
      case 'upcoming': return 'UPCOMING';
      case 'ended': return 'FINAL';
      default: return status.toUpperCase();
    }
  };

  const formatGameTime = (game: UpcomingGame) => {
    const gameDate = new Date(game.currentState.clock || '');
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const gameDay = new Date(gameDate.getFullYear(), gameDate.getMonth(), gameDate.getDate());
    
    // Check if game is today
    if (gameDay.getTime() === today.getTime()) {
      return `Today ${gameDate.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })}`;
    }
    
    // Check if game is tomorrow
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);
    if (gameDay.getTime() === tomorrow.getTime()) {
      return `Tomorrow ${gameDate.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })}`;
    }
    
    // Otherwise show day and time
    return gameDate.toLocaleDateString('en-US', { weekday: 'short', hour: 'numeric', minute: '2-digit', hour12: true });
  };

  const isToday = (game: UpcomingGame) => {
    const gameDate = new Date(game.currentState.clock || '');
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const gameDay = new Date(gameDate.getFullYear(), gameDate.getMonth(), gameDate.getDate());
    return gameDay.getTime() === today.getTime();
  };

  const formatGameInfo = (game: Game) => {
    const { currentState } = game;
    
    // Show period/inning information
    if (currentState.clock) {
      return currentState.clock;
    }
    
    // Show time remaining (count, etc.)
    if (currentState.timeRemaining) {
      return currentState.timeRemaining;
    }
    
    // Show venue if available
    if (currentState.venue?.name) {
      return currentState.venue.name;
    }
    
    return '';
  };

  const renderTeamLogo = (team: { teamCode: string; teamName: string; logoUrl?: string }) => {
    if (team.logoUrl) {
      return (
        <img 
          src={team.logoUrl} 
          alt={`${team.teamName} logo`}
          className="w-10 h-10 object-contain"
          onError={(e) => {
            // Fallback to team code if logo fails to load
            e.currentTarget.style.display = 'none';
            e.currentTarget.nextElementSibling?.classList.remove('hidden');
          }}
        />
      );
    }
    return (
      <div className="w-10 h-10 bg-gray-600 rounded-full flex items-center justify-center text-sm font-bold text-white">
        {team.teamCode}
      </div>
    );
  };

  // Baseball Diamond Component
  const BaseballDiamond: React.FC<{ 
    baseRunners?: { first?: any; second?: any; third?: any };
    outs?: number;
    inning?: number;
    isTopInning?: boolean;
  }> = ({ baseRunners, outs = 0, inning, isTopInning }) => {
    return (
      <div className="flex flex-col items-center space-y-2">
        {/* Inning Display */}
        {inning && (
          <div className="text-center">
            <div className="text-xs text-gray-400 uppercase tracking-wide">
              {isTopInning ? 'Top' : 'Bottom'} {inning}
            </div>
          </div>
        )}
        
        {/* Diamond */}
        <div className="relative w-24 h-24">
          {/* Diamond outline */}
          <svg className="absolute inset-0 w-full h-full" viewBox="0 0 100 100">
            <path
              d="M50 10 L90 50 L50 90 L10 50 Z"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              className="text-gray-400"
            />
            {/* Home plate */}
            <polygon
              points="50,90 45,85 55,85"
              fill="currentColor"
              className="text-gray-400"
            />
          </svg>
          
          {/* Base runners */}
          {baseRunners?.first && (
            <div className="absolute top-1 left-1/2 transform -translate-x-1/2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
              1
            </div>
          )}
          {baseRunners?.second && (
            <div className="absolute top-1/2 right-1 transform translate-y-1/2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
              2
            </div>
          )}
          {baseRunners?.third && (
            <div className="absolute bottom-1 left-1/2 transform translate-x-1/2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
              3
            </div>
          )}
        </div>
        
        {/* Outs */}
        {outs > 0 && (
          <div className="flex space-x-1">
            {[...Array(3)].map((_, i) => (
              <div
                key={i}
                className={`w-3 h-3 rounded-full ${
                  i < outs ? 'bg-red-500' : 'bg-gray-600'
                }`}
              />
            ))}
          </div>
        )}
      </div>
    );
  };

  // Pitcher/Batter Info Component
  const PitcherBatterInfo: React.FC<{ 
    pitcher?: { name: string; number: number };
    batter?: { name: string; number: number };
    count?: { balls: number; strikes: number };
  }> = ({ pitcher, batter, count }) => {
    return (
      <div className="space-y-2 text-sm">
        {pitcher && (
          <div className="flex justify-between">
            <span className="text-gray-400">P:</span>
            <span className="text-white font-medium">
              #{pitcher.number} {pitcher.name}
            </span>
          </div>
        )}
        {batter && (
          <div className="flex justify-between">
            <span className="text-gray-400">B:</span>
            <span className="text-white font-medium">
              #{batter.number} {batter.name}
            </span>
          </div>
        )}
        {count && (
          <div className="flex justify-between">
            <span className="text-gray-400">Count:</span>
            <span className="text-white font-bold">
              {count.balls}-{count.strikes}
            </span>
          </div>
        )}
      </div>
    );
  };

  if (games.length === 0 && upcomingGames.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 text-lg mb-2">No games scheduled</div>
        <div className="text-gray-500 text-sm">Games will appear here when they start</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {games.map((game) => (
        <div
          key={game.gameCode}
          className="bg-white/10 backdrop-blur-sm rounded-xl border border-white/20 p-6 hover:bg-white/15 transition-all duration-300"
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center space-x-3">
              <span className="text-2xl">{getLeagueIcon(game.leagueId)}</span>
              <div>
                <div className="text-sm text-gray-300 uppercase tracking-wide">
                  {game.currentState.venue?.city && game.currentState.venue?.name 
                    ? `${game.currentState.venue.city} - ${game.currentState.venue.name}`
                    : 'Game'
                  }
                </div>
                <div className="text-xs text-gray-400">
                  {formatGameInfo(game)}
                </div>
              </div>
            </div>
            <div className={`px-3 py-1 rounded-full text-xs font-semibold text-white ${getStatusColor(game.currentState.status)}`}>
              {getStatusText(game.currentState.status)}
            </div>
          </div>

          {/* Main Game Content */}
          <div className="grid grid-cols-12 gap-6 items-center">
            {/* Away Team */}
            <div className="col-span-4 text-right">
              <div className="flex items-center justify-end space-x-3 mb-2">
                {renderTeamLogo(game.currentState.away.team)}
                <div>
                  <div className="text-lg font-semibold text-white">
                    {game.currentState.away.team.teamCode}
                  </div>
                  <div className="text-sm text-gray-300">
                    {game.currentState.away.team.teamName}
                  </div>
                </div>
              </div>
              <div className="text-4xl font-bold text-white">
                {game.currentState.away.score}
              </div>
            </div>

            {/* Center - Baseball Diamond or VS */}
            <div className="col-span-4 flex justify-center">
              {game.leagueId === 2 && game.currentState.status === 'active' ? (
                <BaseballDiamond 
                  baseRunners={game.currentState.details?.baseRunners}
                  outs={game.currentState.details?.outs}
                  inning={game.currentState.period}
                  isTopInning={game.currentState.clock?.includes('Top')}
                />
              ) : (
                <div className="text-gray-400 text-lg font-medium">VS</div>
              )}
            </div>

            {/* Home Team */}
            <div className="col-span-4 text-left">
              <div className="flex items-center space-x-3 mb-2">
                <div>
                  <div className="text-lg font-semibold text-white">
                    {game.currentState.home.team.teamCode}
                  </div>
                  <div className="text-sm text-gray-300">
                    {game.currentState.home.team.teamName}
                  </div>
                </div>
                {renderTeamLogo(game.currentState.home.team)}
              </div>
              <div className="text-4xl font-bold text-white">
                {game.currentState.home.score}
              </div>
            </div>
          </div>

          {/* Baseball-specific details */}
          {game.leagueId === 2 && game.currentState.status === 'active' && (
            <div className="mt-6 pt-4 border-t border-white/10">
              <div className="grid grid-cols-3 gap-6">
                {/* Pitcher/Batter Info */}
                <div className="col-span-1">
                  <PitcherBatterInfo 
                    pitcher={game.currentState.details?.pitcher}
                    batter={game.currentState.details?.batter}
                    count={{
                      balls: game.currentState.details?.ballCount || 0,
                      strikes: game.currentState.details?.strikeCount || 0
                    }}
                  />
                </div>
                
                {/* Game Status */}
                <div className="col-span-1 text-center">
                  <div className="text-sm text-gray-400 mb-1">Game Status</div>
                  <div className="text-white font-medium">
                    {game.currentState.timeRemaining || 'In Progress'}
                  </div>
                </div>
                
                {/* Additional Stats */}
                <div className="col-span-1 text-right">
                  <div className="text-sm text-gray-400 mb-1">Stats</div>
                  <div className="text-white text-sm">
                    {game.currentState.statistics && (
                      <>
                        <div>Hits: {game.currentState.statistics.hits || 0}</div>
                        <div>Errors: {game.currentState.statistics.errors || 0}</div>
                      </>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* General Game Details */}
          {game.currentState.status === 'active' && (
            <div className="mt-4 pt-4 border-t border-white/10">
              <div className="flex justify-between text-sm text-gray-400">
                {game.currentState.period && game.currentState.period > 0 && (
                  <div>
                    <span className="font-medium">Period:</span> {game.currentState.period}
                    {game.currentState.periodType && (
                      <span className="ml-1 text-xs">({game.currentState.periodType})</span>
                    )}
                  </div>
                )}
                {game.currentState.timeRemaining && (
                  <div>
                    <span className="font-medium">Time:</span> {game.currentState.timeRemaining}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      ))}

      {/* Upcoming Games Section */}
      {upcomingGames.length > 0 && (
        <div className="mt-8">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-xl font-bold text-white">Upcoming Games (Next 7 Days)</h3>
            <button
              onClick={() => setShowUpcoming(!showUpcoming)}
              className="text-blue-400 hover:text-blue-300 text-sm"
            >
              {showUpcoming ? 'Hide' : 'Show'} ({upcomingGames.length})
            </button>
          </div>
          
          {showUpcoming && (
            <div className="space-y-3">
              {upcomingGames.slice((currentPage - 1) * gamesPerPage, currentPage * gamesPerPage).map((game) => (
                <div
                  key={game.gameCode}
                  className={`backdrop-blur-sm rounded-lg border p-4 ${
                    isToday(game) 
                      ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 border-blue-400/30 shadow-lg' 
                      : 'bg-white/5 border-white/10'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <span className="text-lg">{getLeagueIcon(game.leagueId)}</span>
                      <div className="flex items-center space-x-2">
                        {renderTeamLogo(game.currentState.away.team)}
                        <span className="text-sm font-medium text-white">
                          {game.currentState.away.team.teamCode}
                        </span>
                      </div>
                      <span className="text-gray-400 text-sm">vs</span>
                      <div className="flex items-center space-x-2">
                        {renderTeamLogo(game.currentState.home.team)}
                        <span className="text-sm font-medium text-white">
                          {game.currentState.home.team.teamCode}
                        </span>
                      </div>
                      {isToday(game) && (
                        <span className="px-2 py-1 bg-blue-500/30 text-blue-300 text-xs rounded-full font-semibold">
                          TODAY
                        </span>
                      )}
                    </div>
                    <div className={`text-xs ${isToday(game) ? 'text-blue-300 font-semibold' : 'text-gray-400'}`}>
                      {formatGameTime(game)}
                    </div>
                  </div>
                </div>
              ))}
              
              {/* Pagination */}
              {upcomingGames.length > gamesPerPage && (
                <div className="flex items-center justify-center space-x-2 mt-4">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="px-3 py-1 text-sm bg-white/10 hover:bg-white/20 disabled:opacity-50 disabled:cursor-not-allowed rounded text-white"
                  >
                    Previous
                  </button>
                  <span className="text-sm text-gray-400">
                    Page {currentPage} of {Math.ceil(upcomingGames.length / gamesPerPage)}
                  </span>
                  <button
                    onClick={() => setCurrentPage(Math.min(Math.ceil(upcomingGames.length / gamesPerPage), currentPage + 1))}
                    disabled={currentPage >= Math.ceil(upcomingGames.length / gamesPerPage)}
                    className="px-3 py-1 text-sm bg-white/10 hover:bg-white/20 disabled:opacity-50 disabled:cursor-not-allowed rounded text-white"
                  >
                    Next
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default Scoreboard;