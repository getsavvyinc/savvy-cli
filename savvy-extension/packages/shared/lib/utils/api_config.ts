import type { Config } from './shared-types';

const isDevelopment =
  typeof process !== 'undefined' && (process.env.NODE_ENV === 'development' || process.env.__DEV__ === 'true');
console.log('isDevelopment logLine', isDevelopment);

export const config: Config = {
  dashboardURL: isDevelopment ? 'http://localhost:5173' : 'https://app.getsavvy.so',
  apiURL: isDevelopment ? 'http://localhost:8080' : 'https://api.getsavvy.so',
  tokenKey: '',
};
