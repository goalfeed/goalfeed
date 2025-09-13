import { render, screen, waitFor } from '@testing-library/react';
import Scoreboard from './Scoreboard';
import { apiClient } from '../utils/api';

jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
  },
}));

describe('Scoreboard', () => {
  beforeEach(() => {
    (apiClient.get as jest.Mock).mockResolvedValue({ data: { success: true, data: [] } });
  });

  it('renders empty state when no games', async () => {
    render(<Scoreboard games={[]} />);
    await waitFor(() => expect(screen.getByText('No games scheduled')).toBeInTheDocument());
  });

  it('renders active games', async () => {
    const games: any = [
      {
        gameCode: 'g1',
        leagueId: 1,
        leagueName: 'NHL',
        currentState: {
          home: { team: { teamCode: 'NYR', teamName: 'Rangers' }, score: 2 },
          away: { team: { teamCode: 'BOS', teamName: 'Bruins' }, score: 1 },
          status: 'active',
          fetchedAt: new Date().toISOString(),
        },
        isFetching: false,
        extTimestamp: new Date().toISOString(),
      },
    ];
    render(<Scoreboard games={games} />);
    expect((await screen.findAllByText('NYR')).length).toBeGreaterThan(0);
    expect((await screen.findAllByText('BOS')).length).toBeGreaterThan(0);
  });
});


