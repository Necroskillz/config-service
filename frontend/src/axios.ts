import axios from 'axios';
import type { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { getAccessToken, refreshFn, setAccessToken } from './auth';

/**
 * Subset of AxiosRequestConfig
 */
export type RequestConfig<TData = unknown> = {
  baseURL?: string;
  url?: string;
  method?: 'GET' | 'PUT' | 'PATCH' | 'POST' | 'DELETE' | 'OPTIONS';
  params?: unknown;
  data?: TData | FormData;
  responseType?: 'arraybuffer' | 'blob' | 'document' | 'json' | 'text' | 'stream';
  signal?: AbortSignal;
  headers?: AxiosRequestConfig['headers'];
};

/**
 * Subset of AxiosResponse
 */
export type ResponseConfig<TData = unknown> = {
  data: TData;
  status: number;
  statusText: string;
  headers: AxiosResponse['headers'];
};

export type ResponseErrorConfig<TError = unknown> = AxiosError<TError>;

export const axiosInstance = axios.create({
  baseURL: 'http://localhost:1323/api',
});

axiosInstance.interceptors.request.use((config) => {
  const accessToken = getAccessToken();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  return config;
});

export class HttpError extends Error {
  constructor(message: string, public status: number) {
    super(message);
  }
}

export const client = async <TData, TError = unknown, TVariables = unknown>(
  config: RequestConfig<TVariables>
): Promise<ResponseConfig<TData>> => {
  const requestFn = async (canRetry = true) =>
    axiosInstance
      .request<TData, ResponseConfig<TData>>({
        ...config,
        headers: {
          ...config.headers,
        },
      })
      .catch(async (e: AxiosError<TError>): Promise<ResponseConfig<TData>> => {
        if (e.response?.status === 401 && canRetry) {
          const { accessToken } = await refreshFn();
          setAccessToken(accessToken);
          return await requestFn(false);
        }

        throw new HttpError((e.response?.data as any)?.message || e.message, e.response?.status || 500);
      });

  return requestFn();
};

export default client;
