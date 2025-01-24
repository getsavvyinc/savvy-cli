export type ValueOf<T> = T[keyof T];

export interface ApiError {
  status: number;
  message: string;
  code?: string;
}

export interface Config {
  dashboardURL: string;
  apiURL: string;
  tokenKey: string;
}

export interface BaseStorage<T> {
  get: () => Promise<T>;
  set: (value: T) => Promise<void>;
  subscribe: (callback: () => void) => () => void;
  getSnapshot: () => T;
}
