export const Level = {
    DEBUG: 10,
    INFO: 20,
    WARN: 30,
    ERROR: 40,
} as const;

export type Level = (typeof Level)[keyof typeof Level];

const dump = (prefix: string, obj: any) => {
    if (console) {
        // eslint-disable-next-line no-console
        console.log(prefix, obj);
    }
};
const valid = (checkLevel: Level) => {
    const logLevel = Level[import.meta.env.VITE_APP_LOG_LEVEL];
    return logLevel <= checkLevel;
};

export const Log = {
    debug(msg: any) {
        if (valid(Level.DEBUG)) {
            dump('[DEBUG] ', msg);
        }
    },
    info(msg: any) {
        if (valid(Level.INFO)) {
            dump('[INFO] ', msg);
        }
    },
    warn(msg: any) {
        if (valid(Level.WARN)) {
            dump('[WARN] ', msg);
        }
    },
    error(msg: any, err?: unknown) {
        if (valid(Level.ERROR)) {
            dump('[ERROR] ', msg);
            console.error(err);
        }
    },
};
