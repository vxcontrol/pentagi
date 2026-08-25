import type { ColumnDef } from '@tanstack/react-table';

import { Ellipsis, Plus, ShieldOff, Trash, Users as UsersIcon } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';

import type { Role } from '@/models/info';
import type { User } from '@/models/user';

import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { ErrorState } from '@/components/shared/error-state';
import { LoadingState } from '@/components/shared/loading-state';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { InputPassword } from '@/components/ui/input-password';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useAppForm } from '@/hooks/use-app-form';
import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';
import { formatDate } from '@/lib/utils/format';
import { useUser } from '@/providers/user-provider';

// The list endpoints are paginated; PentAGI installations hold few accounts, so
// one generous page keeps the table simple and still shows everyone.
const PAGE_SIZE = 1000;
const LIST_QUERY = `?page=1&pageSize=${PAGE_SIZE}&type=init`;

interface CreateUserDialogProps {
    onCreated: () => void;
    onOpenChange: (open: boolean) => void;
    open: boolean;
    roles: Role[];
}

interface RolesResponse {
    roles: Role[];
    total: number;
}

interface UsersResponse {
    total: number;
    users: User[];
}

const createUserSchema = z.object({
    mail: z.string().min(1, { message: 'Email is required' }).email({ message: 'Invalid email' }),
    name: z.string().min(1, { message: 'Name is required' }).max(70, { message: 'Name is too long' }),
    password: z.string().min(8, { message: 'Password must be at least 8 characters' }),
    role_id: z.string().min(1, { message: 'Role is required' }),
});

type CreateUserValues = z.infer<typeof createUserSchema>;

const statusVariants: Record<User['status'], 'default' | 'destructive' | 'secondary'> = {
    active: 'default',
    blocked: 'destructive',
    created: 'secondary',
};

function CreateUserDialog({ onCreated, onOpenChange, open, roles }: CreateUserDialogProps) {
    const form = useAppForm<CreateUserValues>({
        defaultValues: { mail: '', name: '', password: '', role_id: '' },
        schema: createUserSchema,
    });
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async (values: CreateUserValues) => {
        setIsSubmitting(true);

        try {
            const response = await api.post<User>('/users/', {
                mail: values.mail,
                name: values.name,
                password: values.password,
                role_id: Number(values.role_id),
                status: 'active',
                type: 'local',
            });

            unwrapApiResponse(response);
            toast.success(`User ${values.mail} created`);
            form.reset();
            onOpenChange(false);
            onCreated();
        } catch (err) {
            toast.error(
                getApiErrorMessage(err, 'Failed to create the user', {
                    403: 'You are not allowed to create users with this role',
                }),
            );
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Dialog
            onOpenChange={onOpenChange}
            open={open}
        >
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Add user</DialogTitle>
                    <DialogDescription>
                        Creates a local account. Accounts signing in through an identity provider appear here on their
                        first login.
                    </DialogDescription>
                </DialogHeader>
                <Form {...form}>
                    <form
                        className="space-y-4"
                        onSubmit={form.handleSubmit(handleSubmit)}
                    >
                        <FormField
                            control={form.control}
                            name="name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Name</FormLabel>
                                    <FormControl>
                                        <Input
                                            placeholder="Jane Doe"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="mail"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Email</FormLabel>
                                    <FormControl>
                                        <Input
                                            placeholder="jane@example.com"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="password"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Password</FormLabel>
                                    <FormControl>
                                        <InputPassword
                                            placeholder="Initial password"
                                            {...field}
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="role_id"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Role</FormLabel>
                                    <Select
                                        onValueChange={field.onChange}
                                        value={field.value}
                                    >
                                        <FormControl>
                                            <SelectTrigger aria-label="Role">
                                                <SelectValue placeholder="Select a role" />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            {roles.map((role) => (
                                                <SelectItem
                                                    key={role.id}
                                                    value={String(role.id)}
                                                >
                                                    {role.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormSubmitButton loading={isSubmitting}>Create</FormSubmitButton>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}

function SettingsUsers() {
    const { authInfo } = useUser();
    const privileges = useMemo(() => authInfo?.privileges ?? [], [authInfo?.privileges]);
    const canView = privileges.includes('users.view');
    const canEdit = privileges.includes('users.edit');
    const canCreate = privileges.includes('users.create');
    const canDelete = privileges.includes('users.delete');
    const currentUserHash = authInfo?.user?.hash;

    const [users, setUsers] = useState<User[]>([]);
    const [roles, setRoles] = useState<Role[]>([]);
    const [isLoading, setIsLoading] = useState(canView);
    const [error, setError] = useState<null | string>(null);
    const [pendingHash, setPendingHash] = useState<null | string>(null);
    const [userToDelete, setUserToDelete] = useState<null | User>(null);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    // Bumped by the retry button and after a user is created, to refetch both lists.
    const [reloadToken, setReloadToken] = useState(0);

    const reload = useCallback(() => {
        setIsLoading(true);
        setError(null);
        setReloadToken((token) => token + 1);
    }, []);

    useEffect(() => {
        if (!canView) {
            return;
        }

        let isCancelled = false;

        const load = async () => {
            try {
                const [usersResponse, rolesResponse] = await Promise.all([
                    api.get<UsersResponse>(`/users/${LIST_QUERY}`),
                    api.get<RolesResponse>(`/roles/${LIST_QUERY}`),
                ]);

                if (isCancelled) {
                    return;
                }

                setUsers(unwrapApiResponse(usersResponse).users ?? []);
                setRoles(unwrapApiResponse(rolesResponse).roles ?? []);
            } catch (err) {
                if (!isCancelled) {
                    setError(getApiErrorMessage(err, 'Failed to load users'));
                }
            } finally {
                if (!isCancelled) {
                    setIsLoading(false);
                }
            }
        };

        void load();

        return () => {
            isCancelled = true;
        };
    }, [canView, reloadToken]);

    const roleName = useCallback(
        (roleId: number) => roles.find((role) => role.id === roleId)?.name ?? `Role #${roleId}`,
        [roles],
    );

    const patchUser = useCallback(
        async (user: User, changes: Partial<Pick<User, 'role_id' | 'status'>>, successMessage: string) => {
            setPendingHash(user.hash);

            try {
                const response = await api.put<User>(`/users/${user.hash}`, { ...user, ...changes });

                unwrapApiResponse(response);
                setUsers((current) =>
                    current.map((item) => (item.hash === user.hash ? { ...item, ...changes } : item)),
                );
                toast.success(successMessage);
            } catch (err) {
                toast.error(
                    getApiErrorMessage(err, 'Failed to update the user', {
                        403: 'You are not allowed to make this change',
                    }),
                );
            } finally {
                setPendingHash(null);
            }
        },
        [],
    );

    const deleteUser = useCallback(async (user: User) => {
        setPendingHash(user.hash);

        try {
            const response = await api.delete<unknown>(`/users/${user.hash}`);

            unwrapApiResponse(response);
            setUsers((current) => current.filter((item) => item.hash !== user.hash));
            toast.success(`User ${user.mail} deleted`);
        } catch (err) {
            toast.error(
                getApiErrorMessage(err, 'Failed to delete the user', {
                    403: 'You are not allowed to delete this user',
                }),
            );
        } finally {
            setPendingHash(null);
            setUserToDelete(null);
        }
    }, []);

    const columns = useMemo<ColumnDef<User>[]>(
        () => [
            {
                accessorKey: 'name',
                cell: ({ row }) => <span className="font-medium">{row.original.name || '—'}</span>,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Name"
                    />
                ),
            },
            {
                accessorKey: 'mail',
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Email"
                    />
                ),
            },
            {
                accessorKey: 'type',
                cell: ({ row }) => (
                    <span className="text-muted-foreground">
                        {row.original.type === 'oauth' ? (row.original.provider ?? 'oauth') : 'local'}
                    </span>
                ),
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Sign-in"
                    />
                ),
            },
            {
                accessorKey: 'role_id',
                cell: ({ row }) => {
                    const user = row.original;
                    // Changing your own role is refused by the API — it would let the
                    // last administrator lock everyone out — so it is not offered here.
                    const isSelf = user.hash === currentUserHash;

                    if (!canEdit || isSelf) {
                        return <Badge variant="outline">{roleName(user.role_id)}</Badge>;
                    }

                    return (
                        <Select
                            disabled={pendingHash === user.hash}
                            onValueChange={(value) =>
                                void patchUser(
                                    user,
                                    { role_id: Number(value) },
                                    `${user.mail} is now ${roleName(Number(value))}`,
                                )
                            }
                            value={String(user.role_id)}
                        >
                            <SelectTrigger
                                aria-label={`Role of ${user.mail}`}
                                className="h-8 w-36"
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {roles.map((role) => (
                                    <SelectItem
                                        key={role.id}
                                        value={String(role.id)}
                                    >
                                        {role.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Role"
                    />
                ),
            },
            {
                accessorKey: 'status',
                cell: ({ row }) => <Badge variant={statusVariants[row.original.status]}>{row.original.status}</Badge>,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Status"
                    />
                ),
            },
            {
                accessorKey: 'created_at',
                cell: ({ row }) => (
                    <span className="text-muted-foreground">{formatDate(new Date(row.original.created_at))}</span>
                ),
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title="Created"
                    />
                ),
            },
            {
                cell: ({ row }) => {
                    const user = row.original;
                    const isSelf = user.hash === currentUserHash;
                    const canBlock = canEdit && !isSelf;
                    const canRemove = canDelete && !isSelf;

                    if (!canBlock && !canRemove) {
                        return null;
                    }

                    return (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label={`Actions for ${user.mail}`}
                                    disabled={pendingHash === user.hash}
                                    size="icon"
                                    variant="ghost"
                                >
                                    <Ellipsis className="size-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                {canBlock && (
                                    <DropdownMenuItem
                                        onSelect={() =>
                                            void patchUser(
                                                user,
                                                { status: user.status === 'blocked' ? 'active' : 'blocked' },
                                                user.status === 'blocked'
                                                    ? `${user.mail} unblocked`
                                                    : `${user.mail} blocked`,
                                            )
                                        }
                                    >
                                        <ShieldOff className="size-4" />
                                        {user.status === 'blocked' ? 'Unblock' : 'Block'}
                                    </DropdownMenuItem>
                                )}
                                {canRemove && (
                                    <DropdownMenuItem
                                        onSelect={() => setUserToDelete(user)}
                                        variant="destructive"
                                    >
                                        <Trash className="size-4" />
                                        Delete
                                    </DropdownMenuItem>
                                )}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    );
                },
                id: 'actions',
            },
        ],
        [canDelete, canEdit, currentUserHash, patchUser, pendingHash, roleName, roles],
    );

    if (!canView) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <AppHeaderTitle>Users</AppHeaderTitle>
                    </AppHeaderContent>
                </AppHeader>
                <Empty className="h-full">
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <UsersIcon />
                        </EmptyMedia>
                        <EmptyTitle>Not available</EmptyTitle>
                        <EmptyDescription>
                            Managing users requires administrator privileges on this account.
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            </>
        );
    }

    return (
        <>
            <AppHeader>
                <AppHeaderContent>
                    <AppHeaderTitle>Users</AppHeaderTitle>
                </AppHeaderContent>
                {canCreate && (
                    <AppHeaderActions>
                        <AppHeaderAction
                            icon={<Plus />}
                            label="Add User"
                            onClick={() => setIsCreateOpen(true)}
                        />
                    </AppHeaderActions>
                )}
            </AppHeader>

            {isLoading && <LoadingState title="Loading users..." />}
            {!isLoading && error && (
                <ErrorState
                    message={error}
                    onRetry={reload}
                    title="Failed to load users"
                />
            )}
            {!isLoading && !error && (
                <div className="p-4">
                    <DataTable
                        columns={columns}
                        data={users}
                        filterColumn="mail"
                        filterPlaceholder="Filter users..."
                    />
                </div>
            )}

            <CreateUserDialog
                onCreated={reload}
                onOpenChange={setIsCreateOpen}
                open={isCreateOpen}
                roles={roles}
            />

            <ConfirmationDialog
                confirmText="Delete"
                description={`${userToDelete?.mail ?? ''} will lose access immediately. This cannot be undone.`}
                handleConfirm={() => (userToDelete ? deleteUser(userToDelete) : undefined)}
                handleOpenChange={(isOpen: boolean) => !isOpen && setUserToDelete(null)}
                isOpen={!!userToDelete}
                title="Delete user?"
            />
        </>
    );
}

export default SettingsUsers;
