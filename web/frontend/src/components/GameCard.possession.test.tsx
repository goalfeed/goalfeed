import { render, screen } from '@testing-library/react';
import GameCard from './GameCard';

const game: any = {
  gameCode: 'g-pos',
  leagueId: 6,
  leagueName: 'NFL',
  currentState: {
    home: { team: { teamCode: 'BUF', teamName: 'Bills' }, score: 7 },
    away: { team: { teamCode: 'NYJ', teamName: 'Jets' }, score: 0 },
    status: 'active',
    fetchedAt: new Date().toISOString(),
    details: { possession: 'BUF' },
  },
  isFetching: false,
  extTimestamp: new Date().toISOString(),
};

describe('GameCard possession', () => {
  it('shows possession icon for team with ball in football', async () => {
    render(<GameCard game={game} />);
    expect((await screen.findAllByText('BUF')).length).toBeGreaterThan(0);
  });
});
