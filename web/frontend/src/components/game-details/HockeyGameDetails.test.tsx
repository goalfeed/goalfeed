import { render, screen } from '@testing-library/react';
import HockeyGameDetails from './HockeyGameDetails';

const game: any = {
  gameCode: 'g1',
  leagueId: 1,
  leagueName: 'NHL',
  currentState: {
    home: { team: { teamCode: 'NYR', teamName: 'Rangers' }, score: 0, periodScores: [1,0] },
    away: { team: { teamCode: 'BOS', teamName: 'Bruins' }, score: 0, periodScores: [0,1] },
    status: 'active',
    fetchedAt: new Date().toISOString(),
    period: 2,
    periodType: 'P',
    clock: '05:12',
    venue: { name: 'MSG', city: 'NYC' },
  },
  isFetching: false,
  extTimestamp: new Date().toISOString(),
};

describe('HockeyGameDetails', () => {
  it('renders period, clock, scores and venue', () => {
    render(<HockeyGameDetails game={game} />);
    expect(screen.getByText(/P 2/)).toBeInTheDocument();
    expect(screen.getByText('05:12')).toBeInTheDocument();
    expect(screen.getByText('MSG')).toBeInTheDocument();
  });
});


