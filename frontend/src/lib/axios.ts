import type { AxiosError } from 'axios';

import Axios from 'axios';

import { AUTH_STORAGE_KEY } from '@/providers/user-provider';

import { Log } from './log';
import { getReturnUrlParam } from './utils/auth';

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
            message: err.message,
            name: err.name,
            response: err.response,
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

                case 401: {
                    Log.warn('Authentication required.');
                    localStorage.removeItem(AUTH_STORAGE_KEY);

                    // Redirect to login with current URL preserved
                    const currentPath = window.location.pathname;

                    if (currentPath !== '/login') {
                        const returnParam = getReturnUrlParam(currentPath);
                        window.location.href = `/login${returnParam}`;
                    }

                    break;
                }

                case 403: {
                    const responseData = err.response?.data as undefined | { code?: string };

                    // Only redirect to login for auth-related 403 errors
                    if (
                        responseData?.code === 'AuthRequired' ||
                        responseData?.code === 'NotPermitted' ||
                        responseData?.code === 'PrivilegesRequired' ||
                        responseData?.code === 'AdminRequired' ||
                        responseData?.code === 'SuperRequired'
                    ) {
                        Log.warn('You do not have permission to execute the api.');
                        localStorage.removeItem(AUTH_STORAGE_KEY);

                        const currentPath = window.location.pathname;

                        if (currentPath !== '/login') {
                            const returnParam = getReturnUrlParam(currentPath);
                            window.location.href = `/login${returnParam}`;
                        }
                    } else {
                        // For other 403 errors (like invalid password), just log
                        Log.warn(err.response?.data);
                    }

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
