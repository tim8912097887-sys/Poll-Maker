import pino, { type LoggerOptions } from 'pino';

// Separate transport config to prevent spread type pollution
const transport =
  process.env.NODE_ENV === 'development'
    ? {
        target: 'pino-pretty',
        options: { colorize: true },
      }
    : undefined;

// Build LoggerOptions safely
const options: LoggerOptions = {
  level: process.env.LOG_LEVEL || 'info',

  base: {
    service: 'poll-service',
  },

  redact: {
    paths: [
      'password',
      'token',
      'accessToken',
      'refreshToken',
      'authorization',
    ],
    censor: '[REDACTED]',
  },

  // IsoTime typed cleanly
  timestamp: pino.stdTimeFunctions.isoTime,

  transport,
};

export const logger = pino(options);
