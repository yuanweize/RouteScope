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

// System Update APIs
export interface SystemInfo {
  version: string;
  commit: string;
  build_date: string;
  go_version: string;
  os: string;
  arch: string;
}

export interface UpdateCheckResult {
  has_update: boolean;
  current_version: string;
  latest_version?: string;
  release_notes?: string;
  release_url?: string;
  published_at?: string;
}

export interface UpdateResult {
  message: string;
  updated: boolean;
  new_version?: string;
  error?: string;
}

export const getSystemInfo = () => request.get<SystemInfo>('/api/v1/system/info');

export const checkUpdate = () => request.get<UpdateCheckResult>('/api/v1/system/check-update');

export const performUpdate = () => request.post<UpdateResult>('/api/v1/system/update');

// Release Assets API
export interface ReleaseAsset {
  name: string;
  download_url: string;
  size: number;
}

export interface ReleasesInfo {
  tag_name: string;
  published_at: string;
  assets: ReleaseAsset[];
}

export const getReleases = () => request.get<ReleasesInfo>('/api/v1/system/releases');

// Database Management API
export interface DatabaseStats {
  size_bytes: number;
  size_human: string;
  record_count: number;
  target_count: number;
  oldest_record?: string;
  newest_record?: string;
  retention_days: number;
}

export interface SystemSettings {
  retention_days: number;
  speed_test_interval_minutes: number;
  ping_interval_seconds: number;
}

export const getDatabaseStats = () => request.get<DatabaseStats>('/api/v1/system/database/stats');

export const cleanDatabase = (days?: number) => 
  request.post<{ message: string; deleted: number }>('/api/v1/system/database/clean', { days });

export const vacuumDatabase = () => 
  request.post<{ message: string }>('/api/v1/system/database/vacuum');

export const getSettings = () => request.get<SystemSettings>('/api/v1/system/settings');

export const saveSettings = (settings: SystemSettings) => 
  request.post<SystemSettings>('/api/v1/system/settings', settings);

// GeoIP Management API
export interface GeoIPStatus {
  available: boolean;
  path: string;
  size_bytes: number;
  size_human: string;
  mod_time?: string;
  last_updated?: string;
  database_type?: string;
  build_epoch?: number;
  build_time?: string;
  ip_version?: number;
  node_count?: number;
  record_size?: number;
  binary_version?: string;
  description?: string;
}

export const getGeoIPStatus = () => request.get<GeoIPStatus>('/api/v1/system/geoip/status');

export const updateGeoIP = () => 
  request.post<{ success: boolean; message: string; size_human?: string }>('/api/v1/system/geoip/update');
