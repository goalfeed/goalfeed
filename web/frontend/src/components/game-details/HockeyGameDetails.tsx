import React from 'react';
import { Game } from '../../types';

interface HockeyGameDetailsProps {
  game: Game;
}

const HockeyGameDetails: React.FC<HockeyGameDetailsProps> = ({ game }) => {
  const { currentState } = game;
  
  return (
    <div className="flex flex-col items-center space-y-3">
      {/* Period and Time */}
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
      
      {/* Period Scores */}
      {currentState.home.periodScores && currentState.home.periodScores.length > 0 && (
        <div className="flex space-x-4 text-sm">
          <div className="text-center">
            <div className="text-gray-400">Period Scores</div>
            <div className="text-white font-bold">
              {currentState.home.periodScores.join('-')} - {currentState.away.periodScores?.join('-') || '0'}
            </div>
          </div>
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

export default HockeyGameDetails;
