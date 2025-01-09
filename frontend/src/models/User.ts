export interface User {
    id: number;
    mail: string;
    name: string;
    created_at: string;
    hash: string;
    password_change_required: boolean;
    provide: string;
    role_id: number;
    status: 'created' | 'active' | 'blocked';
    type: 'local' | 'oauth';
}
