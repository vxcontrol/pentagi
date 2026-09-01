import { describe, expect, it } from 'vitest';

import type { VersionInfoFragmentFragment } from '@/graphql/types';

import { UpdateState } from '@/graphql/types';

import {
    describeTiming,
    describeVersion,
    formatVersionLabel,
    PENTAGI_RELEASES_URL,
    releaseNotesUrl,
    splitVersion,
    UPDATE_INSTRUCTION,
} from './version-status';

function info(overrides: Partial<VersionInfoFragmentFragment> = {}): VersionInfoFragmentFragment {
    return {
        checkedAt: null,
        current: '2.1.0-93e99748',
        failedAt: null,
        latest: null,
        state: UpdateState.UpToDate,
        strategy: 'preview',
        ...overrides,
    };
}

describe('splitVersion', () => {
    it('separates the git revision the release pipeline stamps onto a build', () => {
        expect(splitVersion('2.1.0-93e99748')).toEqual({ revision: '93e99748', version: '2.1.0' });
    });

    it('leaves a prerelease suffix in the version, since it is not a revision', () => {
        expect(splitVersion('2.2.0-rc1')).toEqual({ revision: null, version: '2.2.0-rc1' });
    });

    it('leaves a bare version and a branch name alone', () => {
        expect(splitVersion('2.1.0')).toEqual({ revision: null, version: '2.1.0' });
        expect(splitVersion('develop')).toEqual({ revision: null, version: 'develop' });
    });
});

describe('formatVersionLabel', () => {
    it('prefixes a numeric version with v and leaves a branch name alone', () => {
        expect(formatVersionLabel('2.1.0')).toBe('v2.1.0');
        expect(formatVersionLabel('develop')).toBe('develop');
    });
});

describe('releaseNotesUrl', () => {
    it('points at the tagged release, whether or not the version already carries the v', () => {
        expect(releaseNotesUrl('2.4.0')).toBe(`${PENTAGI_RELEASES_URL}/tag/v2.4.0`);
        expect(releaseNotesUrl('v2.4.0')).toBe(`${PENTAGI_RELEASES_URL}/tag/v2.4.0`);
    });
});

describe('describeVersion', () => {
    it('shows the yellow arrow and tells the operator to run the installer when an update is offered', () => {
        const presentation = describeVersion(info({ latest: '2.4.0', state: UpdateState.UpdateAvailable }));

        expect(presentation.indicator).toBe('update');
        expect(presentation.label).toBe('v2.1.0');
        expect(presentation.revision).toBe('93e99748');
        expect(presentation.latest).toBe('v2.4.0');
        expect(presentation.headline).toBe('New version v2.4.0 is available');
        expect(presentation.tooltip).toBe('New version v2.4.0 is available. Run the installer tool to update.');
        expect(presentation.action).toBe(UPDATE_INSTRUCTION);
    });

    it('shows no indicator when the build is current', () => {
        const presentation = describeVersion(info());

        expect(presentation.indicator).toBe('none');
        expect(presentation.tooltip).toBe('PentAGI v2.1.0 is up to date.');
        expect(presentation.detail).toContain('preview channel');
        expect(presentation.action).toBeNull();
    });

    it.each([
        [UpdateState.Pending, 'Checking for updates…'],
        [UpdateState.Unreachable, 'Update status unknown: the update service could not be reached.'],
        [
            UpdateState.Unknown,
            'Update status unknown: the update service could not tell whether a newer version exists.',
        ],
        [UpdateState.Disabled, 'Update checks are disabled in the server configuration.'],
    ])('shows the yellow question mark for %s', (state, tooltip) => {
        const presentation = describeVersion(info({ state }));

        expect(presentation.indicator).toBe('question');
        expect(presentation.tooltip).toBe(tooltip);
    });

    it('mentions the newest published version when the build could not be matched to one', () => {
        const presentation = describeVersion(info({ latest: '2.4.0', state: UpdateState.Unknown }));

        expect(presentation.indicator).toBe('question');
        expect(presentation.detail).toContain('The newest published version is v2.4.0.');
    });

    it('treats an update that names no version as unknown rather than pointing an arrow at nothing', () => {
        const presentation = describeVersion(info({ latest: null, state: UpdateState.UpdateAvailable }));

        expect(presentation.indicator).toBe('question');
        expect(presentation.action).toBeNull();
    });

    it('treats a state it has never heard of as unknown', () => {
        const presentation = describeVersion(info({ state: 'surprise' as UpdateState }));

        expect(presentation.indicator).toBe('question');
    });
});

describe('describeTiming', () => {
    const now = new Date('2026-09-02T12:00:00Z');

    it('says nothing when nothing has happened yet', () => {
        expect(describeTiming(info(), now)).toEqual([]);
    });

    it('dates the answer and the failure separately', () => {
        expect(
            describeTiming(info({ checkedAt: '2026-09-02T09:00:00Z', failedAt: '2026-09-02T11:55:00Z' }), now),
        ).toEqual(['Checked about 3 hours ago', 'Last check failed 5 minutes ago']);
    });
});
