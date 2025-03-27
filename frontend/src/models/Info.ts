import type { User } from './User';

export type AuthInfoType = 'guest' | 'user';

export interface AuthInfo {
    type: AuthInfoType;
    user?: User;
    providers?: string[];
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
