import type { AxiosError } from 'axios';

import Axios from 'axios';

import { AUTH_STORAGE_KEY } from '@/providers/user-provider';

import { Log } from './log';
import { getReturnUrlParam } from './utils/auth';

/**
 * Gets CSRF token from cookies
 * @param name - The cookie name to retrieve
 * @returns The cookie value or null if not found
 */
function getCookie(name: string): null | string {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);

    if (parts.length === 2) {
        return parts.pop()?.split(';').shift() || null;
    }

    return null;
}

const axios = Axios.create({
    baseURL: '/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true,
    xsrfCookieName: 'XSRF-TOKEN',
    xsrfHeaderName: 'X-XSRF-TOKEN',
});

// Request interceptor to add CSRF token for state-changing operations
axios.interceptors.request.use(
    (config) => {
        // Add CSRF token for state-changing methods (POST, PUT, PATCH, DELETE)
        const method = config.method?.toUpperCase();

        if (method && ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
            // Try to get CSRF token from cookie
            const csrfToken = getCookie('XSRF-TOKEN') || getCookie('csrf_token') || getCookie('csrftoken');

            if (csrfToken) {
                // Add CSRF token to headers
                config.headers['X-XSRF-TOKEN'] = csrfToken;
                config.headers['X-CSRF-Token'] = csrfToken;
            }
        }

        return config;
    },
    (error) => {
        return Promise.reject(error);
    },
);

axios.interceptors.response.use(
    (res) => {
        return res.data;
    },
    (err: AxiosError) => {
        const error = {
            message: err.message,
            name: err.name,
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
                    localStorage.removeItem(AUTH_STORAGE_KEY);

                    // Redirect to login with current URL preserved
                    const currentPath = window.location.pathname;

                    if (currentPath !== '/login') {
                        const returnParam = getReturnUrlParam(currentPath);
                        window.location.href = `/login${returnParam}`;
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
