import request from './request';

export interface Target {
  id?: number;
  name: string;
  address: string;
  desc: string;
  enabled: boolean;
  probe_type: string;
  probe_config: string;
  last_error?: string;
  last_error_at?: string;
}

export interface LogEntry {
  timestamp: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  message: string;
  source?: string;
}

export const login = (username: string, password: string) => {
  return request.post<{ token: string }>('/login', { username, password });
};

export const checkNeedSetup = () => request.get<{ need_setup: boolean }>('/api/v1/need-setup');

export const setupAdmin = (data: { username: string; password: string }) => request.post('/api/v1/setup', data);

export const updatePassword = (newPassword: string) => request.post('/api/v1/user/password', { new_password: newPassword });

export const getTargets = () => request.get<Target[]>('/api/v1/targets');

export const saveTarget = (target: Target) => request.post<Target>('/api/v1/targets', target);

export const deleteTarget = (id: number) => request.delete(`/api/v1/targets/${id}`);

export const getHistory = (params: { target: string; start?: string; end?: string }) => request.get('/api/v1/history', { params });

export const getLatestTrace = (target: string, lang?: string) =>
  request.get('/api/v1/trace', { params: { target, lang } });

export const triggerProbe = (payload?: { target?: string }) => request.post('/api/v1/probe', payload || {});

export const getLogs = (params?: { lines?: number; level?: string }) =>
  request.get<{ logs: LogEntry[]; count: number }>('/api/v1/logs', { params });
