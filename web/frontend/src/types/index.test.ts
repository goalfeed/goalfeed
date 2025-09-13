import { getEventColor, getEventIcon, getEventPriority } from './index';

describe('event utils', () => {
  it('returns icons for key events', () => {
    expect(getEventIcon('goal')).toBeDefined();
    expect(getEventIcon('home_run')).toBeDefined();
    expect(getEventIcon('touchdown')).toBeDefined();
  });
  it('returns colors for event types', () => {
    expect(getEventColor('goal')).toBe('green');
    expect(getEventColor('penalty')).toBe('red');
    expect(getEventColor('game_start')).toBe('blue');
  });
  it('returns priorities for event types', () => {
    expect(getEventPriority('goal')).toBe('high');
    expect(getEventPriority('period_end')).toBe('normal');
  });
});


