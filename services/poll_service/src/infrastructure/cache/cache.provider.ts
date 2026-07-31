import { ConfigService } from '@nestjs/config';
import { LoggerService } from '../configs/logging/logger.service';
import { createClient, RedisClientType } from 'redis';

export const CacheAsyncProvider = 'CacheAsyncProvider';

export const cacheProvider = [
  {
    provide: CacheAsyncProvider,
    inject: [ConfigService, LoggerService],
    useFactory: async (
      configService: ConfigService,
      logger: LoggerService,
    ): Promise<RedisClientType> => {
      const url = configService.get<string>('CACHE_URL');
      const socketTimeout = configService.get<number>(
        'CACHE_SOCKET_TIMEOUT',
        5000,
      );

      const client = createClient({
        url,
        socket: {
          reconnectStrategy: (retries) => {
            const delay = Math.min(retries * 200, 2000);
            logger.warn(`Retrying Redis connection in ${delay}ms`, {
              event: 'redis_retry',
              attempt: retries,
              delay,
            });
            return delay;
          },
          timeout: socketTimeout,
        },
      });

      // Event handlers
      client.on('connect', () => {
        logger.info('Redis connecting...', {
          event: 'redis_status',
          status: 'connecting',
        });
      });

      client.on('ready', () => {
        logger.info('Redis client ready', {
          event: 'redis_status',
          status: 'ready',
        });
      });

      client.on('error', (err) => {
        logger.error('Redis error occurred', {
          event: 'redis_status',
          status: 'error',
          error: err.message,
        });
      });

      client.on('end', () => {
        logger.warn('Redis connection ended', {
          event: 'redis_status',
          status: 'end',
        });
      });

      client.on('reconnecting', () => {
        logger.warn('Redis reconnecting', {
          event: 'redis_status',
          status: 'reconnecting',
        });
      });

      await client.connect();
      return client;
    },
  },
];
