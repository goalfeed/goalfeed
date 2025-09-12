import { render, screen } from '@testing-library/react';
import EventFeed from './EventFeed';
import { Event } from '../types';

const baseEvent: Event = {
  id: '1',
  type: 'goal',
  timestamp: new Date().toISOString(),
  description: 'Scored a goal',
  teamCode: 'NYR',
  teamName: 'Rangers',
  teamHash: 't1',
  leagueId: 1,
  leagueName: 'NHL',
  gameCode: 'g1',
  gameId: 'g1',
  period: 1,
  time: '12:34',
  opponentCode: 'BOS',
  opponentName: 'Bruins',
  opponentHash: 't2',
};

describe('EventFeed', () => {
  it('renders empty state', () => {
    render(<EventFeed events={[]} />);
    expect(screen.getByText('No Recent Events')).toBeInTheDocument();
  });

  it('renders events list with details', () => {
    const events: Event[] = [
      baseEvent,
      { ...baseEvent, id: '2', type: 'penalty', description: '2 min penalty', details: { penaltyType: 'minor', penaltyMinutes: 2 } },
      { ...baseEvent, id: '3', type: 'home_run', leagueId: 2, leagueName: 'MLB', description: 'Home run' },
    ];
    render(<EventFeed events={events} />);
    expect(screen.getByText('Recent Events')).toBeInTheDocument();
    expect(screen.getAllByText(/Rangers/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Bruins/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/goal/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/penalty/i).length).toBeGreaterThan(0);
  });
});


