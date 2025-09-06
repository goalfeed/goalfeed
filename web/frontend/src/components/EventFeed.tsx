import React from 'react';
import { Event, getEventIcon, getEventColor, getEventPriority } from '../types';

interface EventFeedProps {
  events: Event[];
}

const EventFeed: React.FC<EventFeedProps> = ({ events }) => {
  const getLeagueIcon = (leagueId: number) => {
    switch (leagueId) {
      case 1: return '🏒'; // NHL
      case 2: return '⚾'; // MLB
      case 3: return '⚽'; // EPL
      case 4: return '🏒'; // IIHF
      case 5: return '🏈'; // CFL
      default: return '🏆';
    }
  };

  const getLeagueColor = (leagueId: number) => {
    switch (leagueId) {
      case 1: return 'from-blue-500 to-blue-600'; // NHL
      case 2: return 'from-red-500 to-red-600'; // MLB
      case 3: return 'from-green-500 to-green-600'; // EPL
      case 4: return 'from-purple-500 to-purple-600'; // IIHF
      case 5: return 'from-orange-500 to-orange-600'; // CFL
      default: return 'from-gray-500 to-gray-600';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);

    if (diffMins < 1) {
      return 'Just now';
    } else if (diffMins < 60) {
      return `${diffMins}m ago`;
    } else if (diffHours < 24) {
      return `${diffHours}h ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  const formatEventDetails = (event: Event) => {
    const details = [];
    
    if (event.details) {
      // Goal details
      if (event.details.goalType) {
        details.push(event.details.goalType.replace('_', ' '));
      }
      if (event.details.assist1) {
        details.push(`Assist: ${event.details.assist1.name}`);
      }
      
      // Penalty details
      if (event.details.penaltyType) {
        details.push(`${event.details.penaltyType} (${event.details.penaltyMinutes}min)`);
      }
      
      // Baseball details
      if (event.details.inning) {
        details.push(`Inning ${event.details.inning}`);
      }
      if (event.details.outs !== undefined) {
        details.push(`${event.details.outs} outs`);
      }
      if (event.details.bases) {
        details.push(`Bases: ${event.details.bases}`);
      }
      
      // Football details
      if (event.details.down) {
        details.push(`${event.details.down} & ${event.details.distance}`);
      }
      if (event.details.yardLine) {
        details.push(`Yard ${event.details.yardLine}`);
      }
    }
    
    return details;
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'border-red-500/50 bg-red-500/10';
      case 'critical': return 'border-red-600/50 bg-red-600/20';
      case 'normal': return 'border-blue-500/50 bg-blue-500/10';
      case 'low': return 'border-gray-500/50 bg-gray-500/10';
      default: return 'border-blue-500/50 bg-blue-500/10';
    }
  };

  if (events.length === 0) {
    return (
      <div className="p-12 text-center">
        <div className="w-24 h-24 mx-auto mb-6 bg-gradient-to-r from-purple-500 to-pink-600 rounded-full flex items-center justify-center">
          <span className="text-4xl">📰</span>
        </div>
        <h3 className="text-2xl font-bold text-white mb-3">No Recent Events</h3>
        <p className="text-slate-400 text-lg mb-6">
          Events will appear here when goals are scored or games update
        </p>
        <div className="inline-flex items-center px-4 py-2 bg-purple-600/20 text-purple-400 rounded-lg border border-purple-500/30">
          <span className="mr-2">🔔</span>
          Events will show up automatically when they happen
        </div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h2 className="text-3xl font-bold text-white mb-2">Recent Events</h2>
          <p className="text-slate-400">Live updates from your monitored teams</p>
        </div>
        <div className="flex items-center space-x-2 px-4 py-2 bg-purple-500/20 rounded-lg border border-purple-500/30">
          <div className="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></div>
          <span className="text-purple-400 font-medium">{events.length} Events</span>
        </div>
      </div>

      <div className="space-y-4">
        {events.map((event, index) => {
          const priority = getEventPriority(event.type);
          const eventDetails = formatEventDetails(event);
          
          return (
            <div 
              key={`${event.teamCode}-${event.timestamp}-${index}`}
              className={`group bg-gradient-to-r from-white/10 to-white/5 backdrop-blur-sm rounded-2xl border ${getPriorityColor(priority)} hover:border-white/30 transition-all duration-300 hover:scale-[1.02]`}
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <div className="p-6">
                <div className="flex items-start space-x-4">
                  {/* League Icon */}
                  <div className={`w-12 h-12 bg-gradient-to-r ${getLeagueColor(event.leagueId)} rounded-xl flex items-center justify-center flex-shrink-0`}>
                    <span className="text-lg">{getLeagueIcon(event.leagueId)}</span>
                  </div>

                  {/* Event Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <span className={`px-2 py-1 ${getEventColor(event.type) === 'green' ? 'bg-green-500/20 text-green-400' : 
                          getEventColor(event.type) === 'red' ? 'bg-red-500/20 text-red-400' :
                          getEventColor(event.type) === 'yellow' ? 'bg-yellow-500/20 text-yellow-400' :
                          getEventColor(event.type) === 'blue' ? 'bg-blue-500/20 text-blue-400' :
                          'bg-gray-500/20 text-gray-400'} text-xs font-bold rounded-full border`}>
                          {getEventIcon(event.type)} {event.type.toUpperCase()}
                        </span>
                        <span className="text-sm text-slate-400">{event.leagueName}</span>
                        {event.period && (
                          <span className="text-sm text-slate-400">Period {event.period}</span>
                        )}
                      </div>
                      <span className="text-xs text-slate-500">
                        {formatTimestamp(event.timestamp)}
                      </span>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center space-x-3">
                        <div className="w-8 h-8 bg-gradient-to-r from-slate-600 to-slate-700 rounded-lg flex items-center justify-center">
                          <span className="text-xs font-bold text-white">
                            {event.teamCode}
                          </span>
                        </div>
                        <div>
                          <p className="font-semibold text-white text-sm">
                            {event.teamName}
                          </p>
                          <p className="text-xs text-slate-400">{event.description}</p>
                          {event.playerName && (
                            <p className="text-xs text-slate-500">
                              Player: {event.playerName} #{event.playerNumber}
                            </p>
                          )}
                        </div>
                      </div>

                      {/* Event Details */}
                      {eventDetails.length > 0 && (
                        <div className="ml-11">
                          <div className="flex flex-wrap gap-1">
                            {eventDetails.map((detail, idx) => (
                              <span key={idx} className="px-2 py-1 bg-slate-700/50 text-slate-300 text-xs rounded">
                                {detail}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}

                      <div className="flex items-center space-x-2 text-slate-400 ml-11">
                        <span className="text-xs">vs</span>
                        <div className="w-6 h-6 bg-gradient-to-r from-slate-600 to-slate-700 rounded flex items-center justify-center">
                          <span className="text-xs font-bold text-white">
                            {event.opponentCode}
                          </span>
                        </div>
                        <span className="text-sm">{event.opponentName}</span>
                        {event.score && (
                          <span className="text-sm font-medium">
                            ({event.score.awayScore} - {event.score.homeScore})
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Timestamp Badge */}
                  <div className="flex-shrink-0">
                    <div className="w-16 h-16 bg-gradient-to-br from-slate-700/50 to-slate-800/50 rounded-xl flex items-center justify-center border border-slate-600/50">
                      <div className="text-center">
                        <div className="text-xs text-slate-400">Time</div>
                        <div className="text-sm font-bold text-white">
                          {event.time || new Date(event.timestamp).toLocaleTimeString([], { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })}
                        </div>
                        {event.clock && (
                          <div className="text-xs text-slate-500">{event.clock}</div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Bottom Border Animation */}
              <div className="h-1 bg-gradient-to-r from-transparent via-blue-500/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            </div>
          );
        })}
      </div>

      {/* Load More Button */}
      {events.length >= 50 && (
        <div className="mt-8 text-center">
          <button className="px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-lg font-medium transition-all duration-200 hover:scale-105">
            Load More Events
          </button>
        </div>
      )}
    </div>
  );
};

export default EventFeed;