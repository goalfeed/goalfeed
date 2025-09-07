import React from 'react';
import { Game } from '../../types';

interface FootballGameDetailsProps {
  game: Game;
}

const FootballGameDetails: React.FC<FootballGameDetailsProps> = ({ game }) => {
  const { currentState } = game;
  
  return (
    <div className="flex flex-col items-center space-y-3">
      {/* Quarter and Time */}
      <div className="text-center">
        <div className="text-lg font-bold text-white">
          {currentState.periodType} {currentState.period || 1}
        </div>
        {currentState.clock && (
          <div className="text-sm text-gray-300">
            {currentState.clock}
          </div>
        )}
      </div>
      
      {/* Possession */}
      {currentState.details?.possession && (
        <div className="text-center">
          <div className="text-xs text-gray-400 uppercase tracking-wide">Possession</div>
          <div className="text-sm font-bold text-blue-400">
            {currentState.details.possession}
          </div>
        </div>
      )}
      
      {/* Down and Distance */}
      {currentState.details && (currentState.details.down || currentState.details.distance) && (
        <div className="flex space-x-4 text-sm">
          {currentState.details.down && (
            <div className="text-center">
              <div className="text-gray-400">Down</div>
              <div className="text-white font-bold">{currentState.details.down}</div>
            </div>
          )}
          {currentState.details.distance && (
            <div className="text-center">
              <div className="text-gray-400">To Go</div>
              <div className="text-white font-bold">{currentState.details.distance}</div>
            </div>
          )}
          {currentState.details.yardLine && (
            <div className="text-center">
              <div className="text-gray-400">Yard Line</div>
              <div className="text-white font-bold">{currentState.details.yardLine}</div>
            </div>
          )}
        </div>
      )}
      
      {/* Venue */}
      {currentState.venue && (
        <div className="text-center text-sm text-gray-400">
          {currentState.venue.name}
        </div>
      )}
    </div>
  );
};

export default FootballGameDetails;
