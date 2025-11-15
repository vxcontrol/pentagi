import type { User } from './User';

export interface AuthInfo {
    develop?: boolean;
    expires_at?: string;
    issued_at?: string;
    oauth?: boolean;
    privileges?: string[];
    providers?: string[];
    role?: Role;
    type: AuthInfoType;
    user?: User;
}

export interface AuthInfoResponse {
    data?: AuthInfo;
    error?: string;
    status: AuthResponseStatus;
}

export type AuthInfoType = 'guest' | 'user';

export interface AuthLoginResponse {
    data?: unknown;
    error?: string;
    status: AuthResponseStatus;
}

export type AuthResponseStatus = 'error' | 'success';

export interface Role {
    id: number;
    name: string;
}
