import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import Scoreboard from './Scoreboard';
import { apiClient } from '../utils/api';

jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
  },
}));

const makeGame = (overrides: any = {}) => ({
  gameCode: Math.random().toString(36).slice(2),
  leagueId: 1,
  leagueName: 'NHL',
  currentState: {
    home: { team: { teamCode: 'H', teamName: 'Home' }, score: 0 },
    away: { team: { teamCode: 'A', teamName: 'Away' }, score: 0 },
    status: 'upcoming',
  },
  ...overrides,
});

describe('Scoreboard upcoming list', () => {
  beforeEach(() => {
    const today = new Date();
    const tomorrow = new Date();
    tomorrow.setDate(today.getDate() + 1);
    (apiClient.get as jest.Mock).mockImplementation((path: string) => {
      if (path === '/api/upcoming') {
        return Promise.resolve({
          data: {
            success: true,
            data: [
              makeGame({ gameDetails: { gameDate: today.toISOString() }, currentState: { home: { team: { teamCode: 'H', teamName: 'Home' }, score: 0 }, away: { team: { teamCode: 'A', teamName: 'Away' }, score: 0 }, status: 'upcoming' } }),
              makeGame({ gameDetails: { gameDate: tomorrow.toISOString() } }),
              // invalid clock/date fallback
              makeGame({ currentState: { home: { team: { teamCode: 'H', teamName: 'Home' }, score: 0 }, away: { team: { teamCode: 'A', teamName: 'Away' }, score: 0 }, status: 'upcoming', clock: 'TBD' } }),
              // more to trigger pagination
              ...Array.from({ length: 6 }).map(() => makeGame({ gameDetails: { gameDate: tomorrow.toISOString() } })),
            ],
          },
        });
      }
      return Promise.resolve({ data: { success: true, data: [] } });
    });
  });

  it('shows upcoming toggle and paginates list', async () => {
    render(<Scoreboard games={[]} />);
    await waitFor(() => expect(screen.getByText(/Upcoming Games/)).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /Show \(/ });
    fireEvent.click(toggle);
    // page indicators
    expect(screen.getByText(/Page 1 of/)).toBeInTheDocument();
    // go next
    const next = screen.getByRole('button', { name: 'Next' });
    fireEvent.click(next);
    expect(screen.getByText(/Page 2 of/)).toBeInTheDocument();
    const prev = screen.getByRole('button', { name: 'Previous' });
    fireEvent.click(prev);
    expect(screen.getByText(/Page 1 of/)).toBeInTheDocument();
  });
});


