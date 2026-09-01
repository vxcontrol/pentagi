import { formatDistance } from 'date-fns';

import type { VersionInfoFragmentFragment } from '@/graphql/types';

import { UpdateState } from '@/graphql/types';

export const PENTAGI_RELEASES_URL = 'https://github.com/vxcontrol/pentagi/releases';

// What the badge tells the operator to do. Nothing is applied from the UI: the
// installer is the tool that pulls images and restarts the stack, and it lives on
// the host, not behind a button here.
export const UPDATE_INSTRUCTION =
    'Run the installer tool on the host where PentAGI is deployed and choose Maintenance → Update PentAGI.';

export type VersionIndicator = 'none' | 'question' | 'update';

export interface VersionPresentation {
    /** What the operator should do, when there is something to do. */
    action: null | string;
    /** A sentence or two for the popover, under the headline. */
    detail: string;
    /** One line that names the situation. */
    headline: string;
    indicator: VersionIndicator;
    /** The running version as the badge shows it: "v2.1.0", or "develop". */
    label: string;
    /** The newest version the server offers, as the badge shows it. */
    latest: null | string;
    /** Git revision baked into the build, when the version carries one. */
    revision: null | string;
    /** The hover text. Short, and complete on its own. */
    tooltip: string;
}

// A build is "<version>-<git revision>" when the release pipeline stamps one. The
// revision is worth showing, but not inside a badge that should read as a version.
const REVISION_SUFFIX = /^(.+)-([0-9a-f]{7,40})$/i;

/** When the verdict was reached, and when the last attempt failed, as relative times. */
export function describeTiming(info: VersionInfoFragmentFragment, now: Date = new Date()): string[] {
    const lines: string[] = [];

    if (info.checkedAt) {
        lines.push(`Checked ${formatDistance(new Date(info.checkedAt), now, { addSuffix: true })}`);
    }

    if (info.failedAt) {
        lines.push(`Last check failed ${formatDistance(new Date(info.failedAt), now, { addSuffix: true })}`);
    }

    return lines;
}

export function describeVersion(info: VersionInfoFragmentFragment): VersionPresentation {
    const { revision, version } = splitVersion(info.current);
    const label = formatVersionLabel(version);
    const latest = info.latest ? formatVersionLabel(info.latest) : null;
    const base = { action: null, label, latest, revision };

    switch (info.state) {
        case UpdateState.Disabled:
            return {
                ...base,
                detail: 'Automatic update checks are switched off in the server configuration.',
                headline: 'Update checks are disabled',
                indicator: 'question',
                tooltip: 'Update checks are disabled in the server configuration.',
            };
        case UpdateState.Pending:
            return {
                ...base,
                detail: 'The first update check has not completed yet. It runs shortly after the server starts.',
                headline: 'Checking for updates',
                indicator: 'question',
                tooltip: 'Checking for updates…',
            };
        case UpdateState.Unreachable:
            return {
                ...base,
                detail: 'The update service could not be reached, so it is unknown whether a newer version exists.',
                headline: 'Update status unknown',
                indicator: 'question',
                tooltip: 'Update status unknown: the update service could not be reached.',
            };
        case UpdateState.UpdateAvailable:
            if (latest) {
                return {
                    ...base,
                    action: UPDATE_INSTRUCTION,
                    detail: `You are running ${label}. Update to ${latest} to get the latest fixes and features.`,
                    headline: `New version ${latest} is available`,
                    indicator: 'update',
                    tooltip: `New version ${latest} is available. Run the installer tool to update.`,
                };
            }

            // The server never reports an update without naming it, but a schema is
            // a promise about shape, not about content — and an arrow pointing at
            // nothing is not something the operator can act on.
            return describeUnknown(base);
        case UpdateState.UpToDate:
            return {
                ...base,
                detail: `${label} is the newest version on the ${info.strategy} channel.`,
                headline: 'PentAGI is up to date',
                indicator: 'none',
                tooltip: `PentAGI ${label} is up to date.`,
            };
        case UpdateState.Unknown:
        default:
            return describeUnknown(base);
    }
}

/** "2.1.0" reads as "v2.1.0"; a branch name such as "develop" is left alone. */
export function formatVersionLabel(version: string): string {
    return /^\d/.test(version) ? `v${version}` : version;
}

/** Where to read what changed in a release. Tags on GitHub carry a `v` prefix. */
export function releaseNotesUrl(version: string): string {
    const tag = version.startsWith('v') ? version : `v${version}`;

    return `${PENTAGI_RELEASES_URL}/tag/${encodeURIComponent(tag)}`;
}

export function splitVersion(current: string): { revision: null | string; version: string } {
    const match = REVISION_SUFFIX.exec(current);
    const version = match?.[1];
    const revision = match?.[2];

    if (version && revision) {
        return { revision, version };
    }

    return { revision: null, version: current };
}

function describeUnknown(base: Pick<VersionPresentation, 'action' | 'label' | 'latest' | 'revision'>) {
    const newest = base.latest ? ` The newest published version is ${base.latest}.` : '';

    return {
        ...base,
        detail: `The update service could not tell whether a newer version exists for this build.${newest}`,
        headline: 'Update status unknown',
        indicator: 'question' as const,
        tooltip: 'Update status unknown: the update service could not tell whether a newer version exists.',
    };
}
