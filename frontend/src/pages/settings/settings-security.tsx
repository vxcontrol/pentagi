import { AlertTriangle, Info, Shield, ShieldCheck, ShieldOff } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';

import { AppHeader, AppHeaderContent, AppHeaderTitle } from '@/components/layouts/app/app-header';
import { ErrorState } from '@/components/shared/error-state';
import { LoadingState } from '@/components/shared/loading-state';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';

interface DeafGuardConfig {
    enabled: boolean;
    mode: string;
}

const logModeInfo = {
    description: 'Classifies and logs all commands but does not block anything. Use for initial deployment.',
    icon: <Info className="size-4 text-blue-500" />,
    label: 'Log Only',
};

const modeDescriptions: Record<string, typeof logModeInfo> = {
    enforce: {
        description: 'Blocks both BLOCK-tier and WARN-tier commands. Most restrictive mode.',
        icon: <ShieldCheck className="size-4 text-green-500" />,
        label: 'Enforce',
    },
    log: logModeInfo,
    warn: {
        description: 'Blocks BLOCK-tier commands (container escape, destructive ops) but allows WARN-tier through.',
        icon: <AlertTriangle className="size-4 text-yellow-500" />,
        label: 'Warn',
    },
};

const tiers = [
    {
        action: 'BLOCK',
        description: 'mount, docker socket, nsenter, chroot, /proc access',
        name: 'Container Escape',
        tier: 1,
    },
    { action: 'BLOCK', description: 'rm -rf /, shred, dd to devices, mkfs', name: 'Destructive File Ops', tier: 2 },
    {
        action: 'BLOCK',
        description: 'Flood attacks, fork bombs, stress tools, iptables DROP',
        name: 'Network DoS',
        tier: 3,
    },
    {
        action: 'BLOCK',
        description: 'Backdoor users, SSH keys, cron jobs, systemd services',
        name: 'Persistence / Implants',
        tier: 4,
    },
    {
        action: 'WARN',
        description: 'Bash/nc/python reverse shells, curl upload, scp, rsync',
        name: 'Reverse Shells / Exfil',
        tier: 5,
    },
    { action: 'WARN', description: 'hydra, crackmapexec, medusa, ncrack', name: 'Credential Abuse', tier: 6 },
    {
        action: 'WARN',
        description: 'sqlmap --risk=3, nmap exploit scripts, high thread counts',
        name: 'Aggressive Flags',
        tier: 7,
    },
    { action: 'LOG', description: 'nmap, nuclei, curl, ffuf, gobuster, nikto', name: 'Standard Pentesting', tier: 8 },
    { action: 'LOG', description: 'cat, grep, python, jq, find, echo', name: 'Local Utilities', tier: 9 },
];

function SettingsSecurity() {
    const [config, setConfig] = useState<DeafGuardConfig | null>(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [loadError, setLoadError] = useState<string | null>(null);

    const fetchConfig = useCallback(async () => {
        setLoading(true);
        setLoadError(null);
        try {
            const response = await api.get<DeafGuardConfig>('/deafguard/config');
            setConfig(unwrapApiResponse(response));
        } catch (error) {
            const message = getApiErrorMessage(error, 'Failed to load Deaf Guard configuration');
            setLoadError(message);
            toast.error(message);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        void fetchConfig();
    }, [fetchConfig]);

    const updateConfig = useCallback(
        async (updates: Partial<DeafGuardConfig>) => {
            if (!config) {
                return;
            }
            setSaving(true);
            try {
                const response = await api.put<DeafGuardConfig, Partial<DeafGuardConfig>>('/deafguard/config', updates);
                setConfig(unwrapApiResponse(response));
                toast.success('Deaf Guard configuration updated');
            } catch (error) {
                toast.error(getApiErrorMessage(error, 'Failed to update configuration'));
            } finally {
                setSaving(false);
            }
        },
        [config],
    );

    const pageHeader = (
        <AppHeader>
            <AppHeaderContent>
                <AppHeaderTitle icon={<Shield className="size-4 shrink-0" />}>Security</AppHeaderTitle>
            </AppHeaderContent>
        </AppHeader>
    );

    if (loading) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <LoadingState
                        description="Please wait while we fetch the Deaf Guard configuration"
                        title="Loading security settings..."
                    />
                </div>
            </>
        );
    }

    if (!config) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <ErrorState
                        message={loadError}
                        onRetry={fetchConfig}
                        title="Failed to load Deaf Guard configuration"
                    />
                </div>
            </>
        );
    }

    const modeInfo = modeDescriptions[config.mode] ?? logModeInfo;

    return (
        <>
            {pageHeader}
            <div className="mx-auto flex w-full max-w-3xl flex-col gap-4 p-4">
                <Card>
                    <CardHeader className="flex-row items-center gap-4">
                        {config.enabled ? (
                            <Shield className="text-foreground size-5 shrink-0" />
                        ) : (
                            <ShieldOff className="text-muted-foreground size-5 shrink-0" />
                        )}
                        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                            <CardTitle>Deaf Guard</CardTitle>
                            <CardDescription>
                                {config.enabled ? 'Active' : 'Disabled'} — {modeInfo.label}
                            </CardDescription>
                        </div>
                        <Badge variant={config.enabled ? 'default' : 'secondary'}>
                            {config.enabled ? 'Active' : 'Disabled'}
                        </Badge>
                    </CardHeader>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Configuration</CardTitle>
                        <CardDescription>
                            Deaf Guard intercepts terminal commands before execution and classifies them through a
                            9-tier risk taxonomy. Dangerous commands are blocked or flagged based on the enforcement
                            mode.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                                <Label htmlFor="dg-enabled">Enable Deaf Guard</Label>
                                <p className="text-muted-foreground text-sm">
                                    When disabled, all commands pass through without classification.
                                </p>
                            </div>
                            <Switch
                                checked={config.enabled}
                                disabled={saving}
                                id="dg-enabled"
                                onCheckedChange={(checked) => void updateConfig({ enabled: checked })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label>Enforcement Mode</Label>
                            <Select
                                disabled={saving || !config.enabled}
                                onValueChange={(value) => void updateConfig({ mode: value })}
                                value={config.mode}
                            >
                                <SelectTrigger className="w-full">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {Object.entries(modeDescriptions).map(([key, info]) => (
                                        <SelectItem
                                            key={key}
                                            value={key}
                                        >
                                            <div className="flex items-center gap-2">
                                                {info.icon}
                                                <span>{info.label}</span>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <p className="text-muted-foreground text-sm">{modeInfo.description}</p>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Classification Tiers</CardTitle>
                        <CardDescription>
                            Commands are classified into tiers based on risk. The enforcement mode determines which
                            tiers are blocked.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3">
                            {tiers.map((tier) => (
                                <div
                                    className="flex items-start justify-between gap-4 rounded-lg border p-3"
                                    key={tier.tier}
                                >
                                    <div className="min-w-0 flex-1 space-y-1">
                                        <div className="flex items-center gap-2">
                                            <span className="text-muted-foreground text-xs font-medium">
                                                Tier {tier.tier}
                                            </span>
                                            <span className="font-medium">{tier.name}</span>
                                        </div>
                                        <p className="text-muted-foreground text-sm">{tier.description}</p>
                                    </div>
                                    <Badge
                                        variant={
                                            tier.action === 'BLOCK'
                                                ? 'destructive'
                                                : tier.action === 'WARN'
                                                  ? 'secondary'
                                                  : 'outline'
                                        }
                                    >
                                        {tier.action}
                                    </Badge>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-start gap-3">
                            <Info className="text-muted-foreground mt-0.5 size-4 shrink-0" />
                            <div className="text-muted-foreground space-y-1 text-sm">
                                <p>
                                    Configuration changes take effect on the <strong>next flow</strong> that starts.
                                    Active flows continue with their original settings.
                                </p>
                                <p>
                                    Set <code className="bg-muted rounded px-1">DEAF_GUARD_ENABLED</code> and{' '}
                                    <code className="bg-muted rounded px-1">DEAF_GUARD_MODE</code> in your{' '}
                                    <code className="bg-muted rounded px-1">.env</code> file for persistent defaults
                                    across container restarts.
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </>
    );
}

export default SettingsSecurity;
