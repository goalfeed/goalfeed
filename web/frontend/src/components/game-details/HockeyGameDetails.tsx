import React from 'react';
import { Game } from '../../types';

interface HockeyGameDetailsProps {
  game: Game;
}

const formatPeriodDisplay = (period: number | undefined, periodType: string | undefined): string => {
  const periodNum = period || 1;
  
  // For regular periods 1-3, just show the number
  if (periodType === 'REGULAR' && periodNum <= 3) {
    return periodNum.toString();
  }
  
  // For overtime periods, show OT1, OT2, etc.
  if (periodType === 'OVERTIME') {
    // Overtime periods typically start at 4 (OT1), 5 (OT2), etc.
    const otNumber = periodNum > 3 ? periodNum - 3 : periodNum;
    return `OT${otNumber}`;
  }
  
  // For shootout
  if (periodType === 'SHOOTOUT') {
    return 'SO';
  }
  
  // For other period types, show periodType and period
  if (periodType && periodType !== 'REGULAR') {
    return `${periodType} ${periodNum}`;
  }
  
  // Default: just show the period number
  return periodNum.toString();
};

const HockeyGameDetails: React.FC<HockeyGameDetailsProps> = ({ game }) => {
  const { currentState } = game;
  
  return (
    <div className="flex flex-col items-center space-y-3">
      {/* Period and Time */}
      <div className="text-center">
        <div className="text-lg font-bold text-white">
          {formatPeriodDisplay(currentState.period, currentState.periodType)}
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
      
      {/* Shots on Goal */}
      {(currentState.home.statistics?.shots !== undefined || currentState.away.statistics?.shots !== undefined) && (
        <div className="flex space-x-4 text-sm">
          <div className="text-center">
            <div className="text-gray-400">Shots on Goal</div>
            <div className="text-white font-bold">
              {currentState.home.statistics?.shots ?? 0} - {currentState.away.statistics?.shots ?? 0}
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
