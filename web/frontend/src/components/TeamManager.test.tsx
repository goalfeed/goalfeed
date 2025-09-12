import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import TeamManager from './TeamManager';
import { apiClient } from '../utils/api';

jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
  },
}));

describe('TeamManager', () => {
  const onUpdateConfig = jest.fn();
  beforeEach(() => {
    jest.useFakeTimers();
    (apiClient.get as jest.Mock).mockResolvedValue({ data: { success: true, data: [
      { code: 'NYR', name: 'Rangers', location: 'NY', logo: '' },
      { code: 'BOS', name: 'Bruins', location: 'Boston', logo: '' },
    ] } });
  });
  afterEach(() => {
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it('shows leagues and supports editing select all and autosave', async () => {
    render(<TeamManager leagueConfigs={[{ leagueId: 1, leagueName: 'NHL', teams: [] }]} onUpdateConfig={onUpdateConfig} />);
    fireEvent.click(screen.getByText('Edit'));
    await waitFor(() => expect(apiClient.get).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('Select All')).toBeInTheDocument());
    fireEvent.click(screen.getByText('Select All'));
    // trigger debounce
    jest.advanceTimersByTime(1000);
    expect(onUpdateConfig).toHaveBeenCalledWith(1, ['NYR','BOS']);
  });
});


