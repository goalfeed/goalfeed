import { render, screen } from '@testing-library/react';
import LogFeed from './LogFeed';

describe('LogFeed', () => {
  it('renders empty state', () => {
    render(<LogFeed logs={[]} />);
    expect(screen.getByText('No Log Entries')).toBeInTheDocument();
  });

  it('renders log entries list', () => {
    const logs = [
      {
        id: '1',
        type: 'state_change',
        leagueId: 1,
        leagueName: 'NHL',
        teamCode: 'NYR',
        metric: 'score',
        before: 1,
        after: 2,
        timestamp: new Date().toISOString(),
      },
    ];
    render(<LogFeed logs={logs as any} />);
    expect(screen.getByText('Logs')).toBeInTheDocument();
    expect(screen.getByText('NHL')).toBeInTheDocument();
  });
});


