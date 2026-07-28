export type ServerSuccess<T> = {
  state: 'success';
  data: T;
  error: null;
  meta: {
    timestamp: string;
  };
};

export type ServerError = {
  state: 'error';
  data: null;
  error: {
    code: string;
    detail: string;
  };
  meta: {
    timestamp: string;
  };
};

export type ServerResponse<T> = ServerSuccess<T> | ServerError;
