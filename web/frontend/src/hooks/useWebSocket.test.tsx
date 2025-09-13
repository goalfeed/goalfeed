import { renderHook, act } from '@testing-library/react';
import { useWebSocket } from './useWebSocket';

class MockWebSocket {
  onopen: (() => void) | null = null;
  onclose: ((event: any) => void) | null = null;
  onerror: ((error: any) => void) | null = null;
  onmessage: ((event: any) => void) | null = null;
  readyState = 0;
  url: string;

  constructor(url: string) {
    this.url = url;
    setTimeout(() => {
      this.readyState = 1;
      this.onopen && this.onopen();
    }, 0);
  }

  close(code?: number, reason?: string) {
    this.readyState = 3;
    this.onclose && this.onclose({ code: code ?? 1000, reason: reason ?? '' });
  }
}

describe('useWebSocket', () => {
  const originalWebSocket = (global as any).WebSocket;

  beforeEach(() => {
    (global as any).WebSocket = MockWebSocket as any;
    jest.useFakeTimers();
  });

  afterEach(() => {
    (global as any).WebSocket = originalWebSocket;
    jest.useRealTimers();
    jest.clearAllMocks();
  });

  it('connects and sets isConnected to true on open', () => {
    const { result } = renderHook(() => useWebSocket('/ws'));
    act(() => {
      jest.runAllTimers();
    });
    expect(result.current.isConnected).toBe(true);
    expect(result.current.socket).toBeTruthy();
  });

  it('dispatches custom event on message', () => {
    const listener = jest.fn();
    window.addEventListener('websocket-message', listener as EventListener);
    const { result } = renderHook(() => useWebSocket('/ws'));
    act(() => {
      jest.runAllTimers();
    });
    const socket = result.current.socket as any;
    act(() => {
      socket.onmessage && socket.onmessage({ data: JSON.stringify({ type: 'test', data: {} }) });
    });
    expect(listener).toHaveBeenCalled();
    window.removeEventListener('websocket-message', listener as EventListener);
  });

  it('reconnects with backoff on abnormal close', () => {
    const { result } = renderHook(() => useWebSocket('/ws'));
    act(() => {
      jest.runAllTimers();
    });
    const socket = result.current.socket as any;
    act(() => {
      socket.close(1001, 'abnormal');
    });
    // advance timers to trigger reconnect attempt
    act(() => {
      jest.advanceTimersByTime(1000);
    });
    expect(result.current.isConnected).toBe(false);
  });
});


