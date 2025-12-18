import React, { useState, useEffect } from 'react';
import { Game } from '../types';
import { apiClient } from '../utils/api';
import GameCard from './GameCard';

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
  gameDetails?: {
    gameDate: string;
    gameTime: string;
    timezone: string;
  };
}

const Scoreboard: React.FC<ScoreboardProps> = ({ games }) => {
  const [upcomingGames, setUpcomingGames] = useState<UpcomingGame[]>([]);
  const [todaysGames, setTodaysGames] = useState<UpcomingGame[]>([]);
  const [historicalGames, setHistoricalGames] = useState<Game[]>([]);
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [showUpcoming, setShowUpcoming] = useState(false);
  const [showHistorical, setShowHistorical] = useState(false);
  const [isLoadingHistorical, setIsLoadingHistorical] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const gamesPerPage = 5;

  useEffect(() => {
    const fetchUpcomingGames = async () => {
      try {
        const response = await apiClient.get('/api/upcoming');
        if (response.data.success) {
          const allGames = response.data.data || [];
          
          // Separate today's games from upcoming games
          const todayGames: UpcomingGame[] = [];
          const upcomingGamesList: UpcomingGame[] = [];
          
          allGames.forEach((game: UpcomingGame) => {
            if (isToday(game)) {
              todayGames.push(game);
            } else {
              upcomingGamesList.push(game);
            }
          });
          
          setTodaysGames(todayGames);
          setUpcomingGames(upcomingGamesList);
        }
      } catch (error) {
        console.error('Failed to fetch upcoming games:', error);
      }
    };

    fetchUpcomingGames();
  }, []);

  const fetchHistoricalGames = async (date: string) => {
    if (!date) {
      setHistoricalGames([]);
      return;
    }

    setIsLoadingHistorical(true);
    try {
      const response = await apiClient.get(`/api/games/history?date=${date}`);
      if (response.data.success) {
        setHistoricalGames(response.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch historical games:', error);
      setHistoricalGames([]);
    } finally {
      setIsLoadingHistorical(false);
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const date = e.target.value;
    setSelectedDate(date);
    if (date) {
      fetchHistoricalGames(date);
      setShowHistorical(true);
    } else {
      setShowHistorical(false);
      setHistoricalGames([]);
    }
  };

  const getLeagueIcon = (leagueId: number) => {
    switch (leagueId) {
      case 1: return '🏒'; // NHL
      case 2: return '⚾'; // MLB
      case 5: return '🏈'; // CFL
      case 6: return '🏈'; // NFL
      default: return '🏆';
    }
  };

  const formatGameTime = (game: UpcomingGame) => {
    // Try to get the game date from gameDetails if available
    let gameDate: Date;
    
    // Check if we have gameDetails with a proper date
    if (game.gameDetails?.gameDate) {
      gameDate = new Date(game.gameDetails.gameDate);
    } else {
      // Fallback: try to parse clock as a date, but handle invalid dates gracefully
      const clockValue = game.currentState.clock || '';
      gameDate = new Date(clockValue);
      
      // If the date is invalid, return a fallback
      if (isNaN(gameDate.getTime())) {
        return clockValue || 'TBD';
      }
    }
    
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
    // Try to get the game date from gameDetails if available
    let gameDate: Date;
    
    // Check if we have gameDetails with a proper date
    if (game.gameDetails?.gameDate) {
      gameDate = new Date(game.gameDetails.gameDate);
    } else {
      // Fallback: try to parse clock as a date, but handle invalid dates gracefully
      const clockValue = game.currentState.clock || '';
      gameDate = new Date(clockValue);
      
      // If the date is invalid, return false
      if (isNaN(gameDate.getTime())) {
        return false;
      }
    }
    
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const gameDay = new Date(gameDate.getFullYear(), gameDate.getMonth(), gameDate.getDate());
    return gameDay.getTime() === today.getTime();
  };

  const renderTeamLogo = (team: { teamCode: string; teamName: string; logoUrl?: string }) => {
    if (team.logoUrl) {
      return (
        <img 
          src={team.logoUrl} 
          alt={`${team.teamName} logo`}
          className="w-6 h-6 object-contain"
          onError={(e) => {
            // Fallback to team code if logo fails to load
            e.currentTarget.style.display = 'none';
            e.currentTarget.nextElementSibling?.classList.remove('hidden');
          }}
        />
      );
    }
    return (
      <div className="w-6 h-6 bg-gray-600 rounded-full flex items-center justify-center text-xs font-bold text-white">
        {team.teamCode}
      </div>
    );
  };

  if (games.length === 0 && todaysGames.length === 0 && upcomingGames.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 text-lg mb-2">No games scheduled</div>
        <div className="text-gray-500 text-sm">Games will appear here when they start</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Date Picker for Historical Games */}
      <div className="bg-white/5 backdrop-blur-sm rounded-lg border border-white/10 p-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-lg font-bold text-white">Look Up Completed Games</h3>
          {selectedDate && (
            <button
              onClick={() => {
                setSelectedDate('');
                setShowHistorical(false);
                setHistoricalGames([]);
              }}
              className="text-sm text-gray-400 hover:text-white"
            >
              Clear
            </button>
          )}
        </div>
        <div className="flex items-center space-x-3">
          <input
            type="date"
            value={selectedDate}
            onChange={handleDateChange}
            max={new Date().toISOString().split('T')[0]} // Don't allow future dates
            className="px-4 py-2 bg-white/10 border border-white/20 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {isLoadingHistorical && (
            <span className="text-sm text-gray-400">Loading...</span>
          )}
        </div>
      </div>

      {/* Historical Games */}
      {showHistorical && historicalGames.length > 0 && (
        <div className="mt-4">
          <h3 className="text-xl font-bold text-white mb-4">
            Games on {selectedDate ? new Date(selectedDate).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }) : ''}
          </h3>
          <div className="space-y-3">
            {historicalGames.map((game) => (
              <GameCard key={game.gameCode} game={game} />
            ))}
          </div>
        </div>
      )}

      {showHistorical && !isLoadingHistorical && historicalGames.length === 0 && selectedDate && (
        <div className="text-center py-8 bg-white/5 backdrop-blur-sm rounded-lg border border-white/10">
          <div className="text-gray-400 text-lg mb-2">No games found</div>
          <div className="text-gray-500 text-sm">No games were played on this date for your monitored teams</div>
        </div>
      )}

      {/* Debug information for MLB games */}
      {games.filter(game => game.leagueId === 2).length > 0 && (
        <div className="bg-blue-500/20 border border-blue-500/30 rounded-lg p-4">
          <h3 className="text-blue-300 font-bold mb-2">MLB Debug Info</h3>
          {games.filter(game => game.leagueId === 2).map(game => (
            <div key={game.gameCode} className="text-sm text-blue-200">
              <div>Game: {game.gameCode}</div>
              <div>Status: {game.currentState.status}</div>
              <div>Period: {game.currentState.period}</div>
              <div>PeriodType: {game.currentState.periodType}</div>
              <div>Clock: {game.currentState.clock}</div>
              <div>Time Remaining: {game.currentState.timeRemaining}</div>
              <div>Baseball Details:</div>
              <div className="ml-4">
                <div>Inning: {game.currentState.details?.inning || 'N/A'}</div>
                <div>Outs: {game.currentState.details?.outs || 'N/A'}</div>
                <div>Balls: {game.currentState.details?.ballCount || 'N/A'}</div>
                <div>Strikes: {game.currentState.details?.strikeCount || 'N/A'}</div>
                <div>Pitcher: {game.currentState.details?.pitcher?.name || 'N/A'}</div>
                <div>Batter: {game.currentState.details?.batter?.name || 'N/A'}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Active Games */}
      {games.map((game) => (
        <GameCard key={game.gameCode} game={game} />
      ))}

      {/* Today's Games */}
      {todaysGames.length > 0 && (
        <div className="mt-8">
          <h3 className="text-xl font-bold text-white mb-4">Today's Games</h3>
          <div className="space-y-3">
            {todaysGames.map((game) => (
              <div
                key={game.gameCode}
                className="backdrop-blur-sm rounded-lg border p-4 bg-gradient-to-r from-blue-500/20 to-purple-500/20 border-blue-400/30 shadow-lg"
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
                    <span className="px-2 py-1 bg-blue-500/30 text-blue-300 text-xs rounded-full font-semibold">
                      TODAY
                    </span>
                  </div>
                  <div className="text-xs text-blue-300 font-semibold">
                    {formatGameTime(game)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

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