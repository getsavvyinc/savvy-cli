import type { BaseStorage } from './shared-types';

export function createStorage<T>(key: string, defaultValue: T, options?: { liveUpdate?: boolean }): BaseStorage<T> {
  return {
    get: async () => {
      const result = await chrome.storage.local.get(key);
      return result[key] ?? defaultValue;
    },
    set: async (value: T) => {
      await chrome.storage.local.set({ [key]: value });
    },
    subscribe: callback => {
      if (options?.liveUpdate) {
        const listener = (changes: { [key: string]: chrome.storage.StorageChange }) => {
          if (key in changes) {
            callback();
          }
        };
        chrome.storage.local.onChanged.addListener(listener);
        return () => chrome.storage.local.onChanged.removeListener(listener);
      }
      return () => {};
    },
    getSnapshot: () => {
      const value = chrome.storage.local.get(key);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      return (value as any)[key] ?? defaultValue;
    },
  };
}
