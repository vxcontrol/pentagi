export const Level = {
    DEBUG: 10,
    ERROR: 40,
    INFO: 20,
    WARN: 30,
} as const;

export type Level = (typeof Level)[keyof typeof Level];

const dump = (prefix: string, obj: unknown) => {
    if (console) {
        console.log(prefix, obj);
    }
};

const valid = (checkLevel: Level) => {
    const logLevel = Level[import.meta.env.VITE_APP_LOG_LEVEL];

    return logLevel <= checkLevel;
};

export const Log = {
    debug(msg: unknown) {
        if (valid(Level.DEBUG)) {
            dump('[DEBUG] ', msg);
        }
    },
    error(msg: unknown, err?: unknown) {
        if (valid(Level.ERROR)) {
            dump('[ERROR] ', msg);
            console.error(err);
        }
    },
    info(msg: unknown) {
        if (valid(Level.INFO)) {
            dump('[INFO] ', msg);
        }
    },
    warn(msg: unknown) {
        if (valid(Level.WARN)) {
            dump('[WARN] ', msg);
        }
    },
};
