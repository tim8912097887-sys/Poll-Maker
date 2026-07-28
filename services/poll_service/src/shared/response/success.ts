import type { ServerSuccess } from './types.js';

export function successResponse<T>(data: T): ServerSuccess<T> {
  return {
    state: 'success',
    data,
    error: null,
    meta: {
      timestamp: new Date().toISOString(),
    },
  };
}
