import { render, screen } from '@testing-library/react';
import FootballGameDetails from './FootballGameDetails';

const game: any = {
  gameCode: 'g1',
  leagueId: 6,
  leagueName: 'NFL',
  currentState: {
    home: { team: { teamCode: 'BUF', teamName: 'Bills' }, score: 0 },
    away: { team: { teamCode: 'NYJ', teamName: 'Jets' }, score: 0 },
    status: 'active',
    fetchedAt: new Date().toISOString(),
    period: 2,
    periodType: 'Q',
    clock: '10:21',
    details: { down: 2, distance: 8, yardLine: 30, possession: 'BUF' },
  },
  isFetching: false,
  extTimestamp: new Date().toISOString(),
};

describe('FootballGameDetails', () => {
  it('renders quarter, clock and situation', () => {
    render(<FootballGameDetails game={game} />);
    expect(screen.getByText(/Q 2/)).toBeInTheDocument();
    expect(screen.getByText('10:21')).toBeInTheDocument();
    expect(screen.getByText(/2nd & 8/)).toBeInTheDocument();
    expect(screen.getByText(/at BUF 30/)).toBeInTheDocument();
    expect(screen.getByText('BUF')).toBeInTheDocument();
  });
});


