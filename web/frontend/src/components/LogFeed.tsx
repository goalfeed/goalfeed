import React from 'react';
import { AppLogEntry, Event } from '../types';

interface LogFeedProps {
  logs: AppLogEntry[];
}

const LogFeed: React.FC<LogFeedProps> = ({ logs }) => {
  if (!logs || logs.length === 0) {
    return (
      <div className="p-12 text-center">
        <div className="w-24 h-24 mx-auto mb-6 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
          <span className="text-4xl">📜</span>
        </div>
        <h3 className="text-2xl font-bold text-white mb-3">No Log Entries</h3>
        <p className="text-slate-400 text-lg">Logs will appear here as events and state changes occur.</p>
      </div>
    );
  }

  const formatTime = (iso: string) => {
    try {
      return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    } catch {
      return iso;
    }
  };

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-3xl font-bold text-white">Logs</h2>
        <span className="text-slate-400">{logs.length} entries</span>
      </div>
      <div className="space-y-3">
        {logs.map((entry) => (
          <div key={entry.id} className="bg-white/5 border border-white/10 rounded-xl p-4">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <div className="text-slate-300 text-sm">
                  <span className="font-semibold mr-2">{entry.leagueName}</span>
                  <span className="px-2 py-0.5 text-xs rounded bg-slate-700/50 border border-slate-600/50 mr-2">{entry.type}</span>
                  <span className="px-2 py-0.5 text-xs rounded bg-slate-700/50 border border-slate-600/50">{entry.teamCode}</span>
                </div>
                {entry.type === 'state_change' && (
                  <div className="text-white">
                    <span className="text-slate-400 mr-2">{entry.metric}</span>
                    <span className="mr-2">{String(entry.before ?? 'unknown')}</span>
                    <span className="text-slate-500 mr-2">→</span>
                    <span className="font-semibold">{String(entry.after)}</span>
                    {entry.opponent && (
                      <span className="text-slate-500 ml-3">vs {entry.opponent}</span>
                    )}
                  </div>
                )}
                {entry.type === 'event' && entry.event && (
                  <div className="text-white">
                    <span className="font-semibold mr-2">{(entry.event as Event).teamCode}</span>
                    <span className="text-slate-300">{(entry.event as Event).description}</span>
                  </div>
                )}
              </div>
              <div className="text-xs text-slate-500 ml-4 whitespace-nowrap">{formatTime(entry.timestamp)}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default LogFeed;



