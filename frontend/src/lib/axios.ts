import type { AxiosError } from 'axios';
import Axios from 'axios';

import { Log } from './log';

const axios = Axios.create({
    baseURL: '/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true,
});
axios.interceptors.response.use(
    (res) => {
        return res.data;
    },
    (err: AxiosError) => {
        const error = {
            name: err.name,
            message: err.message,
            stack: err.stack,
            statusCode: err.response?.status,
            statusText: err.response?.statusText,
            warnings: undefined,
        };
        if (error.statusCode) {
            Log.warn(`[${error.statusCode}] ${error.statusText || 'empty statusText'}`);
            switch (error.statusCode) {
                case 0: {
                    Log.error('No host was found to connect to.');
                    break;
                }
                case 200: {
                    Log.error(
                        'Failed to parse the return value, please check if the response is returned in JSON format',
                    );
                    break;
                }
                case 400: {
                    if (err.response?.data) {
                        Log.warn(err.response.data);
                        const warns = err.response.data as Record<string, string[]>;
                        const globalMessage = warns[''] || ['Please confirm your input.'];
                        error.message = globalMessage[0] as string;
                    }
                    break;
                }
                case 401:
                case 403: {
                    Log.warn('You do not have permission to execute the api.');
                    localStorage.removeItem('auth');
                    break;
                }
                default: {
                    Log.error(err.response?.data);
                }
            }
        } else {
            Log.error(err);
        }
        return Promise.reject(error);
    },
);
export { axios };
