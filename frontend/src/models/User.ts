export interface User {
    created_at: string;
    hash: string;
    id: number;
    mail: string;
    name: string;
    password_change_required: boolean;
    provide: string;
    role_id: number;
    status: 'active' | 'blocked' | 'created';
    type: 'local' | 'oauth';
}
