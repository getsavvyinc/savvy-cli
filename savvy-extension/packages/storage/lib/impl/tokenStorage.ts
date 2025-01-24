import { createStorage } from '../base/base';

export const tokenStorage = createStorage<string>('savvy_user_key', '', {
  liveUpdate: true,
});
