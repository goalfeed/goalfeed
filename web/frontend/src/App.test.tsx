import { render, screen, waitFor } from '@testing-library/react';
import App from './App';
import { apiClient } from './utils/api';

jest.mock('./utils/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}));

describe('App', () => {
  beforeEach(() => {
    (apiClient.get as jest.Mock)
      .mockResolvedValueOnce({ data: { success: true, data: [] } })
      .mockResolvedValueOnce({ data: { success: true, data: [] } })
      .mockResolvedValueOnce({ data: { success: true, data: [] } })
      .mockResolvedValueOnce({ data: { success: true, data: [] } });
  });

  it('renders header and tabs after data load', async () => {
    render(<App />);
    await waitFor(() => expect(screen.getByText('Goalfeed')).toBeInTheDocument());
    expect(screen.getByText('Live Scoreboard')).toBeInTheDocument();
    expect(screen.getByText('Manage Teams')).toBeInTheDocument();
    expect(screen.getByText('Recent Events')).toBeInTheDocument();
    expect(screen.getByText('Home Assistant')).toBeInTheDocument();
  });
});


