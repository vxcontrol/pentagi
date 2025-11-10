import type { User } from './User';

export type AuthInfoType = 'guest' | 'user';

export interface Role {
    id: number;
    name: string;
}

export interface AuthInfo {
    type: AuthInfoType;
    develop?: boolean;
    user?: User;
    role?: Role;
    providers?: string[];
    privileges?: string[];
    oauth?: boolean;
    issued_at?: string;
    expires_at?: string;
}

export type AuthResponseStatus = 'success' | 'error';

export interface AuthInfoResponse {
    status: AuthResponseStatus;
    data?: AuthInfo;
    error?: string;
}

export interface AuthLoginResponse {
    status: AuthResponseStatus;
    data?: unknown;
    error?: string;
}
