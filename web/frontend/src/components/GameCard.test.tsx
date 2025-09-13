import { render, screen } from '@testing-library/react';
import GameCard from './GameCard';

const baseGame: any = {
  gameCode: 'g1',
  leagueId: 1,
  leagueName: 'NHL',
  currentState: {
    home: { team: { teamCode: 'NYR', teamName: 'Rangers' }, score: 0 },
    away: { team: { teamCode: 'BOS', teamName: 'Bruins' }, score: 0 },
    status: 'active',
    fetchedAt: new Date().toISOString(),
    venue: { name: 'MSG', city: 'NYC' },
  },
  isFetching: false,
  extTimestamp: new Date().toISOString(),
};

describe('GameCard', () => {
  it('renders team codes and status', () => {
    render(<GameCard game={baseGame} />);
    expect(screen.getAllByText('NYR').length).toBeGreaterThan(0);
    expect(screen.getAllByText('BOS').length).toBeGreaterThan(0);
    expect(screen.getByText('LIVE')).toBeInTheDocument();
  });

  it('shows baseball details for MLB league', () => {
    const mlb = {
      ...baseGame,
      leagueId: 2,
      currentState: {
        ...baseGame.currentState,
        details: { outs: 1, ballCount: 2, strikeCount: 1, baseRunners: {} },
        period: 3,
        clock: 'Top 3rd',
      },
    };
    render(<GameCard game={mlb as any} />);
    expect(screen.getByText(/Top 3rd/)).toBeInTheDocument();
  });
});


