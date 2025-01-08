import type { User } from './User';

export type AuthInfoType = 'guest' | 'user';

export interface AuthInfo {
    type: AuthInfoType;
    user?: User;
    providers?: string[];
}

export type AuthInfoResponseStatus = 'success' | 'error';

export interface AuthInfoResponse {
    status: AuthInfoResponseStatus;
    data?: AuthInfo;
    error?: string;
}
