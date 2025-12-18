import React from 'react';
import { Game } from '../../types';

interface FootballGameDetailsProps {
  game: Game;
}

const formatPeriodDisplay = (period: number | undefined, periodType: string | undefined): string => {
  const periodNum = period || 1;
  
  // For regular periods 1-4 (quarters), just show the number
  if (periodType === 'REGULAR' && periodNum <= 4) {
    return periodNum.toString();
  }
  
  // For overtime periods, show OT1, OT2, etc.
  if (periodType === 'OVERTIME') {
    // Overtime periods typically start at 5 (OT1), 6 (OT2), etc.
    const otNumber = periodNum > 4 ? periodNum - 4 : periodNum;
    return `OT${otNumber}`;
  }
  
  // For other period types, show periodType and period
  if (periodType && periodType !== 'REGULAR') {
    return `${periodType} ${periodNum}`;
  }
  
  // Default: just show the period number
  return periodNum.toString();
};

const FootballGameDetails: React.FC<FootballGameDetailsProps> = ({ game }) => {
  const { currentState } = game;
  
  return (
    <div className="flex flex-col items-center space-y-3">
      {/* Quarter and Time */}
      <div className="text-center">
        <div className="text-lg font-bold text-white">
          {formatPeriodDisplay(currentState.period, currentState.periodType)}
        </div>
        {currentState.clock && (
          <div className="text-sm text-gray-300">
            {currentState.clock}
          </div>
        )}
        {/* Compact situation line */}
        {currentState.details && (
          <div className="mt-1 text-xs text-gray-300">
            {(() => {
              const d = currentState.details;
              const parts: string[] = [];
              // e.g., "2nd & 8"
              if (d.down && d.distance) {
                const order = ['1st','2nd','3rd','4th'][Math.max(0, Math.min(3, (d.down as number) - 1))];
                parts.push(`${order} & ${d.distance}`);
              }
              // e.g., "at BUF 30" or fallback "BUF ball at 30"
              if (d.yardLine) {
                if (d.possession) {
                  parts.push(`at ${d.possession} ${d.yardLine}`);
                } else {
                  parts.push(`at ${d.yardLine}`);
                }
              } else if (d.possession) {
                parts.push(`${d.possession} ball`);
              }
              const text = parts.join(' ');
              return text || null;
            })()}
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
