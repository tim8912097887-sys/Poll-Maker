import { ServerError } from './types.js';

export function errorResponse(code: string, detail: string): ServerError {
  return {
    state: 'error',
    data: null,
    error: {
      code,
      detail,
    },
    meta: {
      timestamp: new Date().toISOString(),
    },
  };
}
