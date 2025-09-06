import React, { useState, useCallback } from 'react';
import { LeagueConfig } from '../types';
import { apiClient } from '../utils/api';

interface TeamManagerProps {
  leagueConfigs: LeagueConfig[];
  onUpdateConfig: (leagueId: number, teams: string[]) => void;
}

interface Team {
  code: string;
  name: string;
  location: string;
  logo: string;
}

const TeamManager: React.FC<TeamManagerProps> = ({ leagueConfigs, onUpdateConfig }) => {
  const [editingLeague, setEditingLeague] = useState<number | null>(null);
  const [availableTeams, setAvailableTeams] = useState<{ [key: number]: Team[] }>({});
  const [loadingTeams, setLoadingTeams] = useState<{ [key: number]: boolean }>({});
  const [searchTerms, setSearchTerms] = useState<{ [key: number]: string }>({});
  const [selectedTeams, setSelectedTeams] = useState<{ [key: number]: Set<string> }>({});
  const [autosaveTimeout, setAutosaveTimeout] = useState<{ [key: number]: NodeJS.Timeout | null }>({});

  const getLeagueIcon = (leagueId: number) => {
    switch (leagueId) {
      case 1: return '🏒'; // NHL
      case 2: return '⚾'; // MLB
      case 3: return '⚽'; // EPL
      case 4: return '🏒'; // IIHF
      case 5: return '🏈'; // CFL
      case 6: return '🏈'; // NFL
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
      case 6: return 'from-indigo-500 to-indigo-600'; // NFL
      default: return 'from-gray-500 to-gray-600';
    }
  };

  const fetchAvailableTeams = async (leagueId: number) => {
    if (availableTeams[leagueId]) return; // Already loaded
    
    setLoadingTeams(prev => ({ ...prev, [leagueId]: true }));
    try {
      const response = await apiClient.get(`/api/teams?leagueId=${leagueId}`);
      if (response.data.success) {
        setAvailableTeams(prev => ({ ...prev, [leagueId]: response.data.data }));
      }
    } catch (error) {
      console.error(`Failed to fetch teams for league ${leagueId}:`, error);
    } finally {
      setLoadingTeams(prev => ({ ...prev, [leagueId]: false }));
    }
  };

  const debouncedAutosave = useCallback((leagueId: number, teams: string[]) => {
    // Clear existing timeout
    if (autosaveTimeout[leagueId]) {
      clearTimeout(autosaveTimeout[leagueId]!);
    }

    // Set new timeout
    const timeout = setTimeout(() => {
      onUpdateConfig(leagueId, teams);
      setAutosaveTimeout(prev => ({ ...prev, [leagueId]: null }));
    }, 1000); // 1 second delay

    setAutosaveTimeout(prev => ({ ...prev, [leagueId]: timeout }));
  }, [autosaveTimeout, onUpdateConfig]);

  const handleEditLeague = (league: LeagueConfig) => {
    setEditingLeague(league.leagueId);
    const currentTeams = league.teams || [];
    setSelectedTeams(prev => ({
      ...prev,
      [league.leagueId]: new Set(currentTeams)
    }));
    setSearchTerms(prev => ({ ...prev, [league.leagueId]: '' }));
    
    // Fetch available teams if not already loaded
    fetchAvailableTeams(league.leagueId);
  };

  const handleSaveLeague = (leagueId: number) => {
    const teams = Array.from(selectedTeams[leagueId] || []);
    onUpdateConfig(leagueId, teams);
    setEditingLeague(null);
  };

  const handleCancelEdit = () => {
    setEditingLeague(null);
    setSelectedTeams({});
    setSearchTerms({});
  };

  const handleTeamToggle = (leagueId: number, teamCode: string) => {
    const newSelectedTeams = new Set(selectedTeams[leagueId] || []);
    if (newSelectedTeams.has(teamCode)) {
      newSelectedTeams.delete(teamCode);
    } else {
      newSelectedTeams.add(teamCode);
    }
    
    setSelectedTeams(prev => ({ ...prev, [leagueId]: newSelectedTeams }));
    
    // Trigger autosave
    const teamsArray = Array.from(newSelectedTeams);
    debouncedAutosave(leagueId, teamsArray);
  };

  const handleSelectAll = (leagueId: number) => {
    const allTeams = availableTeams[leagueId] || [];
    const allTeamCodes = allTeams.map(team => team.code);
    setSelectedTeams(prev => ({ ...prev, [leagueId]: new Set(allTeamCodes) }));
    
    // Trigger autosave
    debouncedAutosave(leagueId, allTeamCodes);
  };

  const handleDeselectAll = (leagueId: number) => {
    setSelectedTeams(prev => ({ ...prev, [leagueId]: new Set() }));
    
    // Trigger autosave
    debouncedAutosave(leagueId, []);
  };

  const handleSearchChange = (leagueId: number, value: string) => {
    setSearchTerms(prev => ({ ...prev, [leagueId]: value }));
  };

  const getFilteredTeams = (leagueId: number) => {
    const teams = availableTeams[leagueId] || [];
    const searchTerm = searchTerms[leagueId] || '';
    
    if (!searchTerm) return teams;
    
    return teams.filter(team => 
      team.code.toLowerCase().includes(searchTerm.toLowerCase()) ||
      team.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      team.location.toLowerCase().includes(searchTerm.toLowerCase())
    );
  };

  return (
    <div className="p-8">
      <div className="mb-8">
        <h2 className="text-3xl font-bold text-white mb-2">Team Management</h2>
        <p className="text-slate-400">Configure which teams you want to monitor for live updates</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {leagueConfigs.map((league) => (
          <div 
            key={league.leagueId}
            className="bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-sm rounded-2xl border border-white/20 p-6"
          >
            {/* League Header */}
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center space-x-3">
                <div className={`w-12 h-12 bg-gradient-to-r ${getLeagueColor(league.leagueId)} rounded-xl flex items-center justify-center`}>
                  <span className="text-xl">{getLeagueIcon(league.leagueId)}</span>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-white">{league.leagueName}</h3>
                  <p className="text-sm text-slate-400">
                    {(league.teams || []).length} team{(league.teams || []).length !== 1 ? 's' : ''} monitored
                  </p>
                </div>
              </div>
              <button
                onClick={() => handleEditLeague(league)}
                className="px-4 py-2 bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 rounded-lg border border-blue-500/30 transition-colors"
              >
                Edit
              </button>
            </div>

            {/* Teams List */}
            <div className="space-y-3">
              {(league.teams || []).length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {(league.teams || []).map((team) => (
                    <span
                      key={team}
                      className="px-3 py-1 bg-slate-700/50 text-slate-300 rounded-lg text-sm font-medium"
                    >
                      {team}
                    </span>
                  ))}
                </div>
              ) : (
                <div className="text-center py-6">
                  <div className="text-slate-500 mb-2">No teams configured</div>
                  <p className="text-sm text-slate-600">Add team codes to start monitoring</p>
                </div>
              )}
            </div>

            {/* Edit Mode */}
            {editingLeague === league.leagueId && (
              <div className="mt-6 space-y-4">
                {/* Team Selection */}
                {availableTeams[league.leagueId] && (
                  <div>
                    <div className="flex items-center justify-between mb-3">
                      <label className="text-sm font-medium text-slate-300">
                        Select Teams
                      </label>
                      <div className="flex items-center space-x-3">
                        <div className="text-xs text-slate-400">
                          {(selectedTeams[league.leagueId] || new Set()).size} selected
                        </div>
                        <div className="flex space-x-1">
                          <button
                            onClick={() => handleSelectAll(league.leagueId)}
                            className="px-2 py-1 text-xs bg-green-600/20 text-green-300 rounded hover:bg-green-600/30 transition-colors"
                          >
                            Select All
                          </button>
                          <button
                            onClick={() => handleDeselectAll(league.leagueId)}
                            className="px-2 py-1 text-xs bg-red-600/20 text-red-300 rounded hover:bg-red-600/30 transition-colors"
                          >
                            Deselect All
                          </button>
                        </div>
                      </div>
                    </div>
                    
                    {/* Search */}
                    <input
                      type="text"
                      value={searchTerms[league.leagueId] || ''}
                      onChange={(e) => handleSearchChange(league.leagueId, e.target.value)}
                      placeholder="Search teams..."
                      className="w-full px-3 py-2 bg-slate-800/50 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm mb-3"
                    />

                    {/* Team Grid */}
                    <div className="max-h-64 overflow-y-auto bg-slate-800/30 rounded-lg p-3">
                      {loadingTeams[league.leagueId] ? (
                        <div className="text-center py-4 text-slate-400">Loading teams...</div>
                      ) : (
                        <div className="grid grid-cols-2 gap-2">
                          {getFilteredTeams(league.leagueId).map((team) => (
                            <label
                              key={team.code}
                              className={`flex items-center space-x-3 p-3 rounded-lg cursor-pointer transition-colors ${
                                (selectedTeams[league.leagueId] || new Set()).has(team.code)
                                  ? 'bg-blue-600/30 text-blue-300'
                                  : 'bg-slate-700/30 hover:bg-slate-600/30 text-slate-300'
                              }`}
                            >
                              <input
                                type="checkbox"
                                checked={(selectedTeams[league.leagueId] || new Set()).has(team.code)}
                                onChange={() => handleTeamToggle(league.leagueId, team.code)}
                                className="rounded border-slate-600 bg-slate-800 text-blue-600 focus:ring-blue-500"
                              />
                              <img
                                src={team.logo}
                                alt={`${team.code} logo`}
                                className="w-6 h-6 object-contain"
                                onError={(e) => {
                                  e.currentTarget.style.display = 'none';
                                }}
                              />
                              <div className="flex-1 min-w-0">
                                <div className="text-sm font-medium truncate">{team.code}</div>
                                <div className="text-xs text-slate-400 truncate">{team.location}</div>
                              </div>
                            </label>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {/* Action Buttons */}
                <div className="flex space-x-3">
                  <button
                    onClick={() => handleSaveLeague(league.leagueId)}
                    className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors"
                  >
                    Save Changes
                  </button>
                  <button
                    onClick={handleCancelEdit}
                    className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-lg font-medium transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Help Section */}
      <div className="mt-8 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl border border-blue-500/20 p-6">
        <div className="flex items-start space-x-4">
          <div className="w-8 h-8 bg-blue-500/20 rounded-lg flex items-center justify-center flex-shrink-0">
            <span className="text-blue-400">💡</span>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-white mb-2">How to Add Teams</h4>
            <ul className="text-slate-300 space-y-1 text-sm">
              <li>• Click checkboxes to select teams from the visual list</li>
              <li>• Changes are automatically saved as you select teams (1 second delay)</li>
              <li>• Use the search box to quickly find specific teams</li>
              <li>• Team logos help identify teams visually</li>
              <li>• Changes take effect immediately and start monitoring those teams</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TeamManager;