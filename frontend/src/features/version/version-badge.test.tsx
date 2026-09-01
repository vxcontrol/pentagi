import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { VersionInfoFragmentFragment } from '@/graphql/types';

import { UpdateState } from '@/graphql/types';

const provider = vi.hoisted(() => ({
    value: { isLoading: false, versionInfo: null as null | VersionInfoFragmentFragment },
}));

vi.mock('@/providers/version-info-provider', () => ({ useVersionInfo: () => provider.value }));

import { VersionAttentionDot, VersionBadge } from './version-badge';

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

const badge = () => screen.getByRole('button', { name: /PentAGI v2\.1\.0/ });

describe('VersionBadge', () => {
    beforeEach(() => {
        provider.value = { isLoading: false, versionInfo: null };
    });

    it('renders nothing until the server has said which version it runs', () => {
        const { container } = render(<VersionBadge />);

        expect(container).toBeEmptyDOMElement();
    });

    it('shows the version without an indicator when the build is current', () => {
        provider.value = { isLoading: false, versionInfo: info() };
        render(<VersionBadge />);

        expect(badge()).toHaveTextContent('v2.1.0');
        expect(screen.queryByTestId('version-indicator-update')).not.toBeInTheDocument();
        expect(screen.queryByTestId('version-indicator-question')).not.toBeInTheDocument();
    });

    it('can carry the product name where there is no logo beside it', () => {
        provider.value = { isLoading: false, versionInfo: info() };
        render(<VersionBadge showProductName />);

        expect(badge()).toHaveTextContent('PentAGI v2.1.0');
    });

    it('shows the yellow arrow and the installer hint when an update is offered', async () => {
        const user = userEvent.setup();
        provider.value = {
            isLoading: false,
            versionInfo: info({ latest: '2.4.0', state: UpdateState.UpdateAvailable }),
        };
        render(<VersionBadge />);

        expect(screen.getByTestId('version-indicator-update')).toHaveClass('text-yellow-500');

        await user.hover(badge());

        expect(await screen.findByRole('tooltip')).toHaveTextContent(
            'New version v2.4.0 is available. Run the installer tool to update.',
        );
    });

    it.each([
        [UpdateState.Unreachable, 'the update service could not be reached'],
        [UpdateState.Unknown, 'could not tell whether a newer version exists'],
        [UpdateState.Pending, 'Checking for updates'],
        [UpdateState.Disabled, 'Update checks are disabled'],
    ])('shows the yellow question mark when the verdict is %s', async (state, tooltip) => {
        const user = userEvent.setup();
        provider.value = { isLoading: false, versionInfo: info({ state }) };
        render(<VersionBadge />);

        expect(screen.getByTestId('version-indicator-question')).toHaveClass('text-yellow-500');

        await user.hover(badge());

        expect(await screen.findByRole('tooltip')).toHaveTextContent(tooltip);
    });

    it('opens the details with the build, the channel, the instruction and the release links', async () => {
        const user = userEvent.setup();
        provider.value = {
            isLoading: false,
            versionInfo: info({
                checkedAt: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
                latest: '2.4.0',
                state: UpdateState.UpdateAvailable,
                strategy: 'stable',
            }),
        };
        render(<VersionBadge />);

        await user.click(badge());

        const dialog = await screen.findByRole('dialog');

        expect(dialog).toHaveTextContent('PentAGI v2.1.0');
        expect(dialog).toHaveTextContent('build 93e99748');
        expect(dialog).toHaveTextContent('stable');
        expect(dialog).toHaveTextContent('New version v2.4.0 is available');
        expect(dialog).toHaveTextContent('Run the installer tool on the host where PentAGI is deployed');
        expect(dialog).toHaveTextContent('Checked 10 minutes ago');
        expect(screen.getByRole('link', { name: /Release notes for v2\.4\.0/ })).toHaveAttribute(
            'href',
            'https://github.com/vxcontrol/pentagi/releases/tag/v2.4.0',
        );
        expect(screen.getByRole('link', { name: /All releases/ })).toHaveAttribute(
            'href',
            'https://github.com/vxcontrol/pentagi/releases',
        );
    });

    it('offers no release notes link when there is no release to point at', async () => {
        const user = userEvent.setup();
        provider.value = { isLoading: false, versionInfo: info({ state: UpdateState.Unreachable }) };
        render(<VersionBadge />);

        await user.click(badge());

        expect(await screen.findByRole('dialog')).toHaveTextContent('Update status unknown');
        expect(screen.queryByRole('link', { name: /Release notes/ })).not.toBeInTheDocument();
        expect(screen.getByRole('link', { name: /All releases/ })).toBeInTheDocument();
    });
});

describe('VersionAttentionDot', () => {
    it('appears only when the badge would carry an indicator', () => {
        provider.value = { isLoading: false, versionInfo: info() };
        const { rerender } = render(<VersionAttentionDot />);

        expect(screen.queryByTestId('version-attention-dot')).not.toBeInTheDocument();

        provider.value = {
            isLoading: false,
            versionInfo: info({ latest: '2.4.0', state: UpdateState.UpdateAvailable }),
        };
        rerender(<VersionAttentionDot />);

        expect(screen.getByTestId('version-attention-dot')).toBeInTheDocument();

        provider.value = { isLoading: false, versionInfo: info({ state: UpdateState.Unreachable }) };
        rerender(<VersionAttentionDot />);

        expect(screen.getByTestId('version-attention-dot')).toBeInTheDocument();
    });
});
