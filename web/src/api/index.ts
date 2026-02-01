import request from './request';

export const login = (username: string, password: string) => {
    return request.post<{ token: string }>('/login', { username, password });
};

export const checkNeedSetup = () => {
    return request.get<{ need_setup: boolean }>('/api/v1/need-setup');
};

export const setupAdmin = (data: any) => {
    return request.post('/api/v1/setup', data);
};

export const updatePassword = (newPassword: string) => {
    return request.post('/api/v1/user/password', { new_password: newPassword });
};

export const getStatus = () => {
    return request.get('/api/v1/status');
};

export const getHistory = (params: { target?: string; start?: string; end?: string }) => {
    return request.get('/api/v1/history', { params });
};

export const triggerProbe = (payload?: { target?: string }) => {
    return request.post('/api/v1/probe', payload || {});
};

export interface Target {
    id?: number;
    name: string;
    address: string;
    desc: string;
    enabled: boolean;
    probe_type: string;
    probe_config: string;
}

export const getTargets = () => {
    return request.get<Target[]>('/api/v1/targets');
};

export const saveTarget = (target: Target) => {
    return request.post<Target>('/api/v1/targets', target);
};

export const deleteTarget = (id: number) => {
    return request.delete(`/api/v1/targets/${id}`);
};

export const getLatestTrace = (target: string) => {
    return request.get('/api/v1/trace', { params: { target } });
};
