import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import HASettings from './HASettings';
import { apiClient } from '../utils/api';

jest.mock('../utils/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}));

describe('HASettings', () => {
  beforeEach(() => {
    (apiClient.get as jest.Mock).mockImplementation((path: string) => {
      if (path.includes('/api/homeassistant/config')) {
        return Promise.resolve({ data: { success: true, data: { configured: { tokenSet: true, url: 'http://ha.local' } } } });
      }
      if (path.includes('/api/homeassistant/status')) {
        return Promise.resolve({ data: { success: true, data: { connected: false, source: 'unset', message: 'no token', url: '', tokenSet: false } } });
      }
      return Promise.resolve({ data: { success: true, data: {} } });
    });
    (apiClient.post as jest.Mock).mockResolvedValue({ data: { success: true } });
  });

  it('loads status and allows saving config', async () => {
    render(<HASettings />);
    await waitFor(() => expect(screen.getByText('Home Assistant Integration')).toBeInTheDocument());

    fireEvent.change(screen.getByPlaceholderText('http://homeassistant.local:8123'), { target: { value: 'http://new-ha' } });
    fireEvent.change(screen.getByPlaceholderText(/Long-Lived|configured/i), { target: { value: 'token' } });
    fireEvent.click(screen.getByText('Save Configuration'));

    await waitFor(() => expect(apiClient.post).toHaveBeenCalled());
  });
});


