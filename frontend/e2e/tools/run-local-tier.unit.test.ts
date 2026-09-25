import { execFileSync, spawn, spawnSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

const RUNNER = join(import.meta.dirname, 'run-local-tier.sh');

// A fake that keeps listing what it was told to remove makes every ordering gate below vacuous.
const FAKE_DOCKER = `#!/usr/bin/env bash
filter_pattern() {
    local previous=""
    for arg in "$@"; do
        if [[ "$previous" == "--filter" ]]; then
            printf '%s' "\${arg#name=}"
            return
        fi
        previous="$arg"
    done
}

forget() {
    local file="$1"
    shift
    for name in "$@"; do
        printf 'removed %s\\n' "$name" >> "$DOCKER_LOG"
        grep -vxF "$name" "$file" > "$file.next" || true
        mv "$file.next" "$file"
    done
}

case "$1" in
    build)
        [[ -n "\${BUILD_FAILS:-}" ]] && exit 1
        exit 0
        ;;
    compose) exit 0 ;;
esac

case "$1 $2" in
    "ps -a") sort -u "$SANDBOX_FILE" | grep -E "$(filter_pattern "$@")" || true ;;
    "volume ls") sort -u "$VOLUME_FILE" | grep -E "$(filter_pattern "$@")" || true ;;
    "rm -f")
        shift 2
        if [[ -n "\${RM_FAILS:-}" ]]; then
            for name in "$@"; do printf 'FAILED %s\n' "$name" >> "$DOCKER_LOG"; done
            exit 1
        fi
        forget "$SANDBOX_FILE" "$@"
        ;;
    "volume rm")
        shift 2
        if [[ -n "\${VOLUME_RM_FAILS:-}" ]]; then
            for name in "$@"; do printf 'FAILED %s\n' "$name" >> "$DOCKER_LOG"; done
            exit 1
        fi
        forget "$VOLUME_FILE" "$@"
        ;;
esac
`;

const HARNESS = `
set -euo pipefail
eval "$(sed -n '/^sandbox_names()/,/^}/p;/^sandbox_volumes()/,/^}/p;/^remove_owned_leftovers()/,/^}/p;/^prepare_sandbox_baseline()/,/^}/p;/^record_owned_sandboxes()/,/^}/p;/^remove_e2e_sandboxes()/,/^}/p;/^cleanup()/,/^}/p' "$RUNNER")"
compose() { :; }
`;

let work: string;

const paths = () => ({
    log: join(work, 'docker.log'),
    sandboxes: join(work, 'sandboxes'),
    state: join(work, '.tier2-sandboxes'),
    volumes: join(work, 'volumes'),
    volumeState: join(work, '.tier2-sandbox-volumes'),
});

const run = (
    script: string,
    {
        isContainerRemovalFailing = false,
        isStackKept = false,
        isVolumeRemovalFailing = false,
        sandboxes = '',
        volumes = '',
    }: {
        isContainerRemovalFailing?: boolean;
        isStackKept?: boolean;
        isVolumeRemovalFailing?: boolean;
        sandboxes?: string;
        volumes?: string;
    },
) => {
    const files = paths();

    writeFileSync(files.sandboxes, sandboxes);
    writeFileSync(files.volumes, volumes);
    writeFileSync(files.log, '');

    execFileSync(
        'bash',
        ['-c', `${HARNESS}OWNED_STATE="${files.state}"\nOWNED_VOLUME_STATE="${files.volumeState}"\n${script}`],
        {
            encoding: 'utf8',
            env: {
                ...process.env,
                DOCKER_LOG: files.log,
                PATH: `${join(work, 'bin')}:${process.env.PATH ?? ''}`,
                RUNNER,
                SANDBOX_FILE: files.sandboxes,
                VOLUME_FILE: files.volumes,
                ...(isVolumeRemovalFailing ? { VOLUME_RM_FAILS: '1' } : {}),
                ...(isContainerRemovalFailing ? { RM_FAILS: '1' } : {}),
                ...(isStackKept ? { E2E_KEEP_STACK: '1' } : {}),
            },
        },
    );

    return {
        recorded: () => readFileSync(files.state, 'utf8').trim(),
        recordedVolumes: () => readFileSync(files.volumeState, 'utf8').trim(),
        removed: () => readFileSync(files.log, 'utf8'),
    };
};

beforeEach(() => {
    work = mkdtempSync(join(tmpdir(), 'tier2-runner-'));
    mkdirSync(join(work, 'bin'));
    writeFileSync(join(work, 'bin', 'docker'), FAKE_DOCKER);
    chmodSync(join(work, 'bin', 'docker'), 0o755);
});

afterEach(() => {
    rmSync(work, { force: true, recursive: true });
});

const portFunctions = () =>
    execFileSync('sed', ['-n', '/^port_is_taken()/,/^}/p;/^port_for()/,/^}/p', RUNNER], { encoding: 'utf8' });

const portBlock = () =>
    execFileSync('sed', ['-n', '/^port_is_taken()/,/PGVECTOR_LISTEN_PORT=/p', RUNNER], { encoding: 'utf8' });

const runPortFor = (defaultPort: number, named?: number) =>
    spawnSync('bash', ['-c', `${portFunctions()}\nport_for PENTAGI_LISTEN_PORT ${defaultPort}`], {
        encoding: 'utf8',
        env: { ...process.env, ...(named === undefined ? {} : { PENTAGI_LISTEN_PORT: String(named) }) },
    });

const runPortBlock = (named?: number) =>
    spawnSync('bash', ['-c', `set -euo pipefail\n${portBlock()}\necho "$PENTAGI_LISTEN_PORT"`], {
        encoding: 'utf8',
        env: { ...process.env, ...(named === undefined ? {} : { PENTAGI_LISTEN_PORT: String(named) }) },
    });

const listenOn = async (): Promise<{ close: () => void; port: number }> => {
    const { createServer } = await import('node:net');
    const server = createServer();

    await new Promise<void>((resolve) => {
        server.listen(0, '127.0.0.1', () => resolve());
    });

    const address = server.address();

    if (address === null || typeof address === 'string') {
        throw new Error('the probe server did not take a port');
    }

    return { close: () => server.close(), port: address.port };
};

const freePort = async (): Promise<number> => {
    const probe = await listenOn();

    probe.close();

    return probe.port;
};

describe('tier-2 port choice', () => {
    it('moves a taken default to the next free port', async () => {
        const taken = await listenOn();

        try {
            expect(runPortFor(taken.port).stdout).toBe(String(taken.port + 1));
        } finally {
            taken.close();
        }
    });

    it('stops the run when the caller names a port something already serves', async () => {
        const taken = await listenOn();

        try {
            const result = runPortBlock(taken.port);

            expect(result.status).toBe(1);
            expect(result.stderr).toContain(`port ${taken.port} is already serving something`);
            expect(result.stdout).toBe('');
        } finally {
            taken.close();
        }
    });

    it('keeps the port the caller named when it is free', async () => {
        const free = await freePort();

        expect(runPortBlock(free).stdout.trim()).toBe(String(free));
    });
});

describe('tier-2 sandbox bookkeeping', { timeout: 30_000 }, () => {
    const SANDBOXES = 'pentagi-terminal-90001\npentagi-terminal-95000\n';
    const VOLUMES = 'pentagi-terminal-90001-data\npentagi-terminal-95000-data\n';

    it('removes the leftovers a previous run recorded, and nothing else', () => {
        writeFileSync(paths().state, 'pentagi-terminal-90001\n');
        writeFileSync(paths().volumeState, 'pentagi-terminal-90001-data\n');

        const { removed } = run('remove_owned_leftovers', { sandboxes: SANDBOXES, volumes: VOLUMES });

        expect(removed()).toContain('removed pentagi-terminal-90001\n');
        expect(removed()).toContain('removed pentagi-terminal-90001-data\n');
        // A developer stack that once ran a tier-2 seed keeps ids in the same range; removing those
        // takes their sandboxes and their data with them.
        expect(removed()).not.toContain('95000');
    });

    it('records what this run added, so the next one can clear it', () => {
        const { recorded } = run(
            'prepare_sandbox_baseline\nprintf "pentagi-terminal-90001\\n" >> "$SANDBOX_FILE"\nrecord_owned_sandboxes',
            { sandboxes: 'pentagi-terminal-95000\n', volumes: 'pentagi-terminal-95000-data\n' },
        );

        expect(recorded()).toBe('pentagi-terminal-90001');
    });

    // The container and the volume are removed by two separate commands, and only one of them may
    // fail. Deriving the volume record from the container names loses the survivor.
    it('keeps a volume whose removal failed, without re-arming the container kill list', () => {
        const { recorded, recordedVolumes, removed } = run(
            [
                'prepare_sandbox_baseline',
                'printf "pentagi-terminal-90001\\n" >> "$SANDBOX_FILE"',
                'printf "pentagi-terminal-90001-data\\n" >> "$VOLUME_FILE"',
                'REPO_ROOT="$(dirname "$OWNED_STATE")" cleanup',
            ].join('\n'),
            {
                isVolumeRemovalFailing: true,
                sandboxes: 'pentagi-terminal-95000\n',
                volumes: 'pentagi-terminal-95000-data\n',
            },
        );

        expect(removed()).toContain('removed pentagi-terminal-90001\n');
        expect(removed()).toContain('FAILED pentagi-terminal-90001-data\n');
        // The container is gone, so nothing may hunt that name again; the volume is still there and
        // stays on the books until someone takes it.
        expect(recorded()).toBe('');
        expect(recordedVolumes()).toBe('pentagi-terminal-90001-data');
    });

    it('leaves nothing in the record once the teardown has taken it', () => {
        const { recorded, removed } = run(
            [
                'prepare_sandbox_baseline',
                'printf "pentagi-terminal-90001\\n" >> "$SANDBOX_FILE"',
                'printf "pentagi-terminal-90001-data\\n" >> "$VOLUME_FILE"',
                'REPO_ROOT="$(dirname "$OWNED_STATE")" cleanup',
            ].join('\n'),
            { sandboxes: 'pentagi-terminal-95000\n', volumes: 'pentagi-terminal-95000-data\n' },
        );

        expect(removed()).toContain('removed pentagi-terminal-90001\n');
        // A name the teardown already took is not this run's to hand to the next one: that file is
        // read as a kill list before anything is built.
        expect(recorded()).toBe('');
    });

    // `set -u` does not catch this: the baseline is read inside a process substitution, and the
    // subshell that dies there still hands `comm` an empty list.
    it('refuses to sweep when the run died before the baseline was taken', () => {
        const { removed } = run('REPO_ROOT="$(dirname "$OWNED_STATE")" cleanup', {
            sandboxes: SANDBOXES,
            volumes: VOLUMES,
        });

        expect(removed()).toBe('');
        // Nor may it hand the next run a kill list it never earned.
        expect(existsSync(paths().state)).toBe(false);
    });

    it('does not take a container outside the seeded id range for its own', () => {
        const { recorded } = run(
            [
                'prepare_sandbox_baseline',
                'printf "pentagi-terminal-42\\npentagi-db\\n" >> "$SANDBOX_FILE"',
                'record_owned_sandboxes',
            ].join('\n'),
            { sandboxes: '', volumes: '' },
        );

        expect(recorded()).toBe('');
    });

    it('keeps a leftover it could not remove on the books for the next run', () => {
        writeFileSync(paths().state, 'pentagi-terminal-90001\n');
        writeFileSync(paths().volumeState, 'pentagi-terminal-90001-data\n');

        const { recorded, recordedVolumes } = run('prepare_sandbox_baseline\nrecord_owned_sandboxes', {
            isContainerRemovalFailing: true,
            isVolumeRemovalFailing: true,
            sandboxes: 'pentagi-terminal-90001\n',
            volumes: 'pentagi-terminal-90001-data\n',
        });

        expect(recorded()).toBe('pentagi-terminal-90001');
        expect(recordedVolumes()).toBe('pentagi-terminal-90001-data');
    });

    it('records the sandboxes of a stack it was told to keep', () => {
        const { recorded, removed } = run(
            [
                'prepare_sandbox_baseline',
                'printf "pentagi-terminal-90001\\n" >> "$SANDBOX_FILE"',
                'REPO_ROOT="$(dirname "$OWNED_STATE")" cleanup',
            ].join('\n'),
            { isStackKept: true, sandboxes: 'pentagi-terminal-95000\n', volumes: 'pentagi-terminal-95000-data\n' },
        );

        expect(removed()).toBe('');
        expect(recorded()).toBe('pentagi-terminal-90001');
    });

    it('still owns the sandbox that took the name of the leftover it removed', () => {
        writeFileSync(paths().state, 'pentagi-terminal-90001\n');
        writeFileSync(paths().volumeState, 'pentagi-terminal-90001-data\n');

        const { recorded, removed } = run(
            [
                'prepare_sandbox_baseline',
                'printf "pentagi-terminal-90001\\n" >> "$SANDBOX_FILE"',
                'printf "pentagi-terminal-90001-data\\n" >> "$VOLUME_FILE"',
                'record_owned_sandboxes',
                'remove_e2e_sandboxes',
            ].join('\n'),
            { sandboxes: SANDBOXES, volumes: VOLUMES },
        );

        expect(recorded()).toBe('pentagi-terminal-90001');
        expect(removed().match(/removed pentagi-terminal-90001\n/g)).toHaveLength(2);
        expect(removed().match(/removed pentagi-terminal-90001-data\n/g)).toHaveLength(2);
        expect(removed()).not.toContain('95000');
    });
    it('clears what it created when the run is cut short by a signal', async () => {
        const files = paths();

        writeFileSync(files.sandboxes, 'pentagi-terminal-90001\n');
        writeFileSync(files.volumes, 'pentagi-terminal-90001-data\n');
        writeFileSync(files.log, '');
        writeFileSync(
            join(work, 'bin', 'corepack'),
            `#!/usr/bin/env bash\nprintf 'pentagi-terminal-90002\\n' >> "$SANDBOX_FILE"\nprintf 'e2e started\\n' >> "$DOCKER_LOG"\nsleep 5\n`,
        );
        chmodSync(join(work, 'bin', 'corepack'), 0o755);

        const child = spawn(RUNNER, [], {
            detached: true,
            env: {
                ...process.env,
                DOCKER_LOG: files.log,
                E2E_SKIP_BUILD: '1',
                OWNED_STATE: files.state,
                OWNED_VOLUME_STATE: files.volumeState,
                PATH: `${join(work, 'bin')}:${process.env.PATH ?? ''}`,
                SANDBOX_FILE: files.sandboxes,
                VOLUME_FILE: files.volumes,
            },
        });

        const ended = new Promise((resolve) => child.on('exit', resolve));

        for (let attempt = 0; attempt < 100; attempt += 1) {
            if (readFileSync(files.log, 'utf8').includes('e2e started')) {
                break;
            }

            await new Promise((resolve) => setTimeout(resolve, 50));
        }

        process.kill(-(child.pid as number), 'SIGTERM');
        await ended;

        const log = readFileSync(files.log, 'utf8');

        expect(log).toContain('removed pentagi-terminal-90002');
        expect(log).not.toContain('removed pentagi-terminal-90001');
    });

    it('takes the baseline before a build that fails, so a foreign sandbox survives the run', () => {
        const files = paths();

        writeFileSync(files.sandboxes, 'pentagi-terminal-90001\n');
        writeFileSync(files.volumes, 'pentagi-terminal-90001-data\n');
        writeFileSync(files.log, '');
        writeFileSync(files.state, 'pentagi-terminal-99999\n');
        writeFileSync(files.volumeState, 'pentagi-terminal-99999-data\n');

        const status = spawnSync(RUNNER, [], {
            encoding: 'utf8',
            env: {
                ...process.env,
                BUILD_FAILS: '1',
                DOCKER_LOG: files.log,
                OWNED_STATE: files.state,
                OWNED_VOLUME_STATE: files.volumeState,
                PATH: `${join(work, 'bin')}:${process.env.PATH ?? ''}`,
                SANDBOX_FILE: files.sandboxes,
                VOLUME_FILE: files.volumes,
            },
        });

        expect(status.status).not.toBe(0);
        expect(readFileSync(files.log, 'utf8')).not.toContain('removed pentagi-terminal-90001');
        expect(readFileSync(files.sandboxes, 'utf8')).toContain('pentagi-terminal-90001');
        expect(readFileSync(files.state, 'utf8').trim()).toBe('');
    });
});
