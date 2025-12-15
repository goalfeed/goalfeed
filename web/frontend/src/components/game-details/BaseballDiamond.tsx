import React from 'react';

interface BaseballDiamondProps {
  baseRunners?: { first?: any; second?: any; third?: any };
  outs?: number;
  inning?: number;
  isTopInning?: boolean;
}

const BaseballDiamond: React.FC<BaseballDiamondProps> = ({ 
  baseRunners, 
  outs = 0, 
  inning, 
  isTopInning 
}) => {
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

export default BaseballDiamond;
