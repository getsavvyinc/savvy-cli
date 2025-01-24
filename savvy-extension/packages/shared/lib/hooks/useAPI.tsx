// src/shared/api/hooks.ts
import { useMemo } from 'react';
import type { AxiosInstance } from 'axios';
import axios from 'axios';
interface LocalApiHookResult {
  client: AxiosInstance;
}

export function useLocalClient(): LocalApiHookResult {
  const axiosInstance = useMemo(
    () =>
      axios.create({
        baseURL: 'http://localhost:8765',
        headers: {
          'Content-Type': '*/*',
        },
      }),
    [],
  ); // Empty dependency array means this will only be created once

  return {
    client: axiosInstance,
  };
}
