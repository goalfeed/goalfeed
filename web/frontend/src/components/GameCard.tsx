import React, { useEffect, useRef, useState } from 'react';
import { Game } from '../types';
import FootballGameDetails from './game-details/FootballGameDetails';
import HockeyGameDetails from './game-details/HockeyGameDetails';
import BaseballDiamond from './game-details/BaseballDiamond';
import PitcherBatterInfo from './game-details/PitcherBatterInfo';

interface GameCardProps {
  game: Game;
}

const GameCard: React.FC<GameCardProps> = ({ game }) => {
  const [homeFlash, setHomeFlash] = useState(false);
  const [awayFlash, setAwayFlash] = useState(false);
  const [homeFlashColor, setHomeFlashColor] = useState<'green' | 'red'>('green');
  const [awayFlashColor, setAwayFlashColor] = useState<'green' | 'red'>('green');
  const prevHomeScore = useRef<number>(game.currentState.home.score);
  const prevAwayScore = useRef<number>(game.currentState.away.score);

  // Detect home score change
  useEffect(() => {
    const current = game.currentState.home.score;
    const prev = prevHomeScore.current;
    if (current !== prev) {
      setHomeFlashColor(current > prev ? 'green' : 'red');
      setHomeFlash(true);
      prevHomeScore.current = current;
      const t = setTimeout(() => setHomeFlash(false), 1200);
      return () => clearTimeout(t);
    }
  }, [game.currentState.home.score]);

  // Detect away score change
  useEffect(() => {
    const current = game.currentState.away.score;
    const prev = prevAwayScore.current;
    if (current !== prev) {
      setAwayFlashColor(current > prev ? 'green' : 'red');
      setAwayFlash(true);
      prevAwayScore.current = current;
      const t = setTimeout(() => setAwayFlash(false), 1200);
      return () => clearTimeout(t);
    }
  }, [game.currentState.away.score]);
  const getLeagueIcon = (leagueId: number) => {
    switch (leagueId) {
      case 1: return '🏒'; // NHL
      case 2: return '⚾'; // MLB
      case 5: return '🏈'; // CFL
      case 6: return '🏈'; // NFL
      default: return '🏆';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500';
      case 'upcoming': return 'bg-blue-500';
      case 'delayed': return 'bg-yellow-500';
      case 'ended': return 'bg-gray-500';
      default: return 'bg-gray-400';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return 'LIVE';
      case 'upcoming': return 'UPCOMING';
      case 'delayed': return 'DELAYED';
      case 'ended': return 'FINAL';
      default: return status.toUpperCase();
    }
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

  const renderGameDetails = () => {
    if (game.leagueId === 2 && (game.currentState.status === 'active' || game.currentState.status === 'delayed')) {
      return (
        <BaseballDiamond 
          baseRunners={game.currentState.details?.baseRunners}
          outs={game.currentState.details?.outs}
          inning={game.currentState.period}
          isTopInning={game.currentState.clock?.includes('Top')}
        />
      );
    } else if ((game.leagueId === 5 || game.leagueId === 6) && (game.currentState.status === 'active' || game.currentState.status === 'delayed')) {
      return <FootballGameDetails game={game} />;
    } else if (game.leagueId === 1 && (game.currentState.status === 'active' || game.currentState.status === 'delayed')) {
      return <HockeyGameDetails game={game} />;
    } else {
      return <div className="text-gray-400 text-lg font-medium">VS</div>;
    }
  };

  return (
    <div className="bg-white/10 backdrop-blur-sm rounded-xl border border-white/20 p-6 hover:bg-white/15 transition-all duration-300">
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
          <div className={`text-4xl font-bold transition-colors ${awayFlash ? 'text-white' : 'text-white'}`}>
            <span className={`${awayFlash ? (awayFlashColor === 'green' ? 'bg-green-500/30 ring-2 ring-green-400/40' : 'bg-red-500/30 ring-2 ring-red-400/40') : ''} px-2 rounded transition-all ${awayFlash ? 'animate-pulse' : ''}`}>
              {game.currentState.away.score}
            </span>
          </div>
        </div>

        {/* Center - Game Details */}
        <div className="col-span-4 flex justify-center">
          {renderGameDetails()}
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
          <div className={`text-4xl font-bold transition-colors ${homeFlash ? 'text-white' : 'text-white'}`}>
            <span className={`${homeFlash ? (homeFlashColor === 'green' ? 'bg-green-500/30 ring-2 ring-green-400/40' : 'bg-red-500/30 ring-2 ring-red-400/40') : ''} px-2 rounded transition-all ${homeFlash ? 'animate-pulse' : ''}`}>
              {game.currentState.home.score}
            </span>
          </div>
        </div>
      </div>

      {/* Baseball-specific details */}
      {game.leagueId === 2 && (game.currentState.status === 'active' || game.currentState.status === 'delayed') && (
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
      {(game.currentState.status === 'active' || game.currentState.status === 'delayed') && (
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
  );
};

export default GameCard;
