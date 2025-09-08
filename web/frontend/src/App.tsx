import React, { useState, useEffect } from 'react';
import { Game, Event, LeagueConfig } from './types';
import { apiClient } from './utils/api';
import { useWebSocket } from './hooks/useWebSocket';
import Scoreboard from './components/Scoreboard';
import TeamManager from './components/TeamManager';
import EventFeed from './components/EventFeed';

const App: React.FC = () => {
  const [games, setGames] = useState<Game[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [leagueConfigs, setLeagueConfigs] = useState<LeagueConfig[]>([]);
  const [activeTab, setActiveTab] = useState<'scoreboard' | 'teams' | 'events'>('scoreboard');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { isConnected } = useWebSocket('/ws');

  // Fetch initial data
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const [gamesResponse, leaguesResponse] = await Promise.all([
          apiClient.get('/api/games'),
          apiClient.get('/api/leagues')
        ]);

        if (gamesResponse.data.success) {
          setGames(gamesResponse.data.data || []);
        }
        if (leaguesResponse.data.success) {
          setLeagueConfigs(leaguesResponse.data.data);
        }
      } catch (err) {
        setError('Failed to load data');
        console.error('Error fetching data:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  // Handle WebSocket messages
  useEffect(() => {
    const handleWebSocketMessage = (event: CustomEvent) => {
      const message = event.detail;
      switch (message.type) {
        case 'game_update':
          setGames(prevGames => {
            const exists = prevGames.some(game => game.gameCode === message.data.gameCode);
            if (exists) {
              return prevGames.map(game => 
                game.gameCode === message.data.gameCode ? message.data : game
              );
            }
            return [message.data, ...prevGames];
          });
          break;
        case 'games_list':
          setGames(message.data || []);
          break;
        case 'games':
          // Backward compatibility with older backend message type
          setGames(message.data || []);
          break;
        case 'event':
          setEvents(prevEvents => [message.data, ...prevEvents.slice(0, 49)]);
          break;
      }
    };

    window.addEventListener('websocket-message', handleWebSocketMessage as EventListener);
    
    return () => {
      window.removeEventListener('websocket-message', handleWebSocketMessage as EventListener);
    };
  }, []);

  const updateLeagueConfig = async (leagueId: number, teams: string[]) => {
    try {
      const response = await apiClient.post('/api/leagues', {
        leagueId,
        teams
      });
      if (response.data.success) {
        setLeagueConfigs(prev => prev.map(config => 
          config.leagueId === leagueId ? { ...config, teams } : config
        ));
        // Reload games to pick up new monitoring
        const gamesResponse = await apiClient.get('/api/games');
        if (gamesResponse.data.success) {
          setGames(gamesResponse.data.data || []);
        }
      }
    } catch (error) {
      console.error('Failed to update league config:', error);
    }
  };

  const tabs = [
    { id: 'scoreboard' as const, label: 'Live Scoreboard', icon: '🏒' },
    { id: 'teams' as const, label: 'Manage Teams', icon: '⚙️' },
    { id: 'events' as const, label: 'Recent Events', icon: '📰' }
  ];

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-white mx-auto mb-4"></div>
          <h2 className="text-2xl font-bold text-white mb-2">Loading Goalfeed</h2>
          <p className="text-slate-300">Connecting to live sports data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-6xl mb-4">⚠️</div>
          <h2 className="text-2xl font-bold text-white mb-2">Connection Error</h2>
          <p className="text-slate-300 mb-4">{error}</p>
          <button 
            onClick={() => window.location.reload()}
            className="px-6 py-3 bg-red-600 hover:bg-red-700 text-white rounded-lg font-medium transition-colors"
          >
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      {/* Header */}
      <header className="bg-black/20 backdrop-blur-md border-b border-white/10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-lg">⚽</span>
              </div>
              <div>
                <h1 className="text-2xl font-bold text-white">Goalfeed</h1>
                <p className="text-sm text-slate-300">Live Sports Scoreboard</p>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400'}`}></div>
                <span className="text-sm text-slate-300">
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              <div className="text-sm text-slate-400">
                {games.length} active games
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <nav className="bg-black/10 backdrop-blur-sm border-b border-white/5">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm transition-all duration-200 ${
                  activeTab === tab.id
                    ? 'border-blue-400 text-blue-400'
                    : 'border-transparent text-slate-400 hover:text-slate-300 hover:border-slate-300'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white/5 backdrop-blur-sm rounded-2xl border border-white/10 overflow-hidden">
          {activeTab === 'scoreboard' && (
            <Scoreboard games={games} />
          )}
          {activeTab === 'teams' && (
            <TeamManager 
              leagueConfigs={leagueConfigs}
              onUpdateConfig={updateLeagueConfig}
            />
          )}
          {activeTab === 'events' && (
            <EventFeed events={events} />
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-black/20 backdrop-blur-md border-t border-white/10 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="text-center text-slate-400 text-sm">
            <p>Goalfeed - Real-time sports monitoring • Built with React & Go</p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default App;