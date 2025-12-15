import React from 'react';

interface PitcherBatterInfoProps {
  pitcher?: { name: string; number: number };
  batter?: { name: string; number: number };
  count?: { balls: number; strikes: number };
}

const PitcherBatterInfo: React.FC<PitcherBatterInfoProps> = ({ 
  pitcher, 
  batter, 
  count 
}) => {
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

export default PitcherBatterInfo;
