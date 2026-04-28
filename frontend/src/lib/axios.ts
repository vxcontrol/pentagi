import type { AxiosError, AxiosRequestConfig } from 'axios';

import Axios from 'axios';

import { AUTH_STORAGE_KEY } from '@/providers/user-provider';

import { Log } from './log';
import { getReturnUrlParam } from './utils/auth';

// ── shared API protocol types ────────────────────────────────────────────────
//
// The backend wraps every successful response as { status: 'success', data?: T }
// and every business-level failure as { status: 'error', code?, msg?, error? }.
// HTTP-level failures (4xx/5xx) are rejected by the response interceptor below
// and surface as `ApiHttpError` to the caller.

export interface ApiErrorResponse {
    code?: string;
    error?: string;
    msg?: string;
    status: 'error';
}

/** Shape returned by the response-error interceptor (rejected promise payload). */
export interface ApiHttpError {
    message: string;
    name: string;
    response?: AxiosError['response'];
    stack?: string;
    statusCode?: number;
    statusText?: string;
    warnings?: unknown;
}

export type ApiResponse<T> = ApiErrorResponse | ApiSuccessResponse<T>;

export type ApiResponseStatus = 'error' | 'success';

export interface ApiSuccessResponse<T> {
    data?: T;
    status: 'success';
}

// ── shared instance ──────────────────────────────────────────────────────────

/**
 * Central axios instance for all REST calls in the app.
 *
 * Best practices from https://axios.rest/:
 *   - Sets a sensible default `timeout` to avoid hanging requests
 *     (override per-call via `{ timeout: 0 }` for long uploads).
 *   - Uses `withCredentials` so the session cookie is sent on every request.
 *   - The response interceptor unwraps `res.data` on success — every method
 *     returns the API payload (typed as `ApiResponse<T>` via the helpers below)
 *     instead of the raw `AxiosResponse<T>`.
 */
const axios = Axios.create({
    baseURL: '/api/v1',
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: 30_000,
    withCredentials: true,
});

axios.interceptors.response.use(
    (res) => res.data,
    (err: AxiosError): Promise<never> => {
        const error: ApiHttpError = {
            message: err.message,
            name: err.name,
            response: err.response,
            stack: err.stack,
            statusCode: err.response?.status,
            statusText: err.response?.statusText,
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

                    const currentPath = window.location.pathname;

                    if (currentPath !== '/login') {
                        const returnParam = getReturnUrlParam(currentPath);
                        window.location.href = `/login${returnParam}`;
                    }

                    break;
                }

                case 403: {
                    const responseData = err.response?.data as undefined | { code?: string };

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

// ── typed wrapper ────────────────────────────────────────────────────────────
//
// The wrapper hides the awkward `<unknown, ApiResponse<T>>` cast that's needed
// because the response interceptor returns `res.data` directly. Callers write:
//
//     const response = await api.post<MyData>('/things', body);
//     // response is typed as ApiResponse<MyData>

export const api = {
    delete: <T>(url: string, config?: AxiosRequestConfig) => axios.delete<T, ApiResponse<T>>(url, config),
    get: <T>(url: string, config?: AxiosRequestConfig) => axios.get<T, ApiResponse<T>>(url, config),
    patch: <T, B = unknown>(url: string, body?: B, config?: AxiosRequestConfig) =>
        axios.patch<T, ApiResponse<T>, B>(url, body, config),
    post: <T, B = unknown>(url: string, body?: B, config?: AxiosRequestConfig) =>
        axios.post<T, ApiResponse<T>, B>(url, body, config),
    put: <T, B = unknown>(url: string, body?: B, config?: AxiosRequestConfig) =>
        axios.put<T, ApiResponse<T>, B>(url, body, config),
};

// ── helpers ──────────────────────────────────────────────────────────────────

export const isApiSuccess = <T>(response: ApiResponse<T>): response is ApiSuccessResponse<T> =>
    response.status === 'success';

/** Returns `response.data`, or throws if the API marked the response as an error. */
export const unwrapApiResponse = <T>(response: ApiResponse<T>): T => {
    if (!isApiSuccess(response) || response.data == null) {
        const message = !isApiSuccess(response) ? (response.msg ?? response.error) : undefined;

        throw new Error(message ?? 'Unexpected response from server');
    }

    return response.data;
};

/**
 * Extracts a human-readable message from an unknown error thrown by axios calls.
 *
 * Lookup order:
 *   1. `response.data.msg` — backend-provided message,
 *   2. `statusFallbacks[status]` — caller-provided defaults per HTTP status,
 *   3. `error.message` — generic axios/network message,
 *   4. `fallback` — last-resort string.
 */
export const getApiErrorMessage = (
    error: unknown,
    fallback: string,
    statusFallbacks?: Record<number, string>,
): string => {
    if (!error || typeof error !== 'object') {
        return fallback;
    }

    const err = error as ApiHttpError;
    const status = err.statusCode ?? err.response?.status;
    const responseData = err.response?.data;
    const responseMsg =
        responseData && typeof responseData === 'object' ? (responseData as ApiErrorResponse).msg : undefined;

    if (responseMsg) {
        return responseMsg;
    }

    if (status != null && statusFallbacks?.[status]) {
        return statusFallbacks[status];
    }

    return err.message || fallback;
};

export default axios;
export { axios };
