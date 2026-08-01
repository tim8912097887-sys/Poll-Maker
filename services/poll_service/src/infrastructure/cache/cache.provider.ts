import { ConfigService } from '@nestjs/config';
import { LoggerService } from '../configs/logging/logger.service';
import { createClient, RedisClientType } from 'redis';
import {
  Injectable,
  OnApplicationShutdown,
  OnModuleInit,
} from '@nestjs/common';

export const CacheAsyncProvider = 'CacheAsyncProvider';

@Injectable()
export class CacheService implements OnModuleInit, OnApplicationShutdown {
  public client: RedisClientType;
  constructor(
    private readonly configService: ConfigService,
    private readonly logger: LoggerService,
  ) {
    const url = this.configService.get<string>('CACHE_URL');
    const socketTimeout = configService.get<number>(
      'CACHE_SOCKET_TIMEOUT',
      5000,
    );

    this.client = createClient({
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
    this.client.on('connect', () => {
      this.logger.info('Redis connecting...', {
        event: 'redis_status',
        status: 'connecting',
      });
    });

    this.client.on('ready', () => {
      this.logger.info('Redis this.client ready', {
        event: 'redis_status',
        status: 'ready',
      });
    });

    this.client.on('error', (err) => {
      this.logger.error('Redis error occurred', {
        event: 'redis_status',
        status: 'error',
        error: err.message,
      });
    });

    this.client.on('end', () => {
      this.logger.warn('Redis connection ended', {
        event: 'redis_status',
        status: 'end',
      });
    });

    this.client.on('reconnecting', () => {
      this.logger.warn('Redis reconnecting', {
        event: 'redis_status',
        status: 'reconnecting',
      });
    });
  }

  async onModuleInit() {
    this.logger.info('Connecting to Redis...', {
      event: 'redis_status',
      status: 'connecting',
    });
    await this.client.connect();

    this.logger.info('Redis connected successfully', {
      event: 'redis_status',
      status: 'connected',
    });
  }

  async onApplicationShutdown(signal?: string) {
    this.logger.info(`Closing Redis due to signal: ${signal}`, {
      event: 'redis_status',
      status: 'disconnecting',
    });

    await this.client.quit();

    this.logger.info('Redis closed gracefully', {
      event: 'redis_status',
      status: 'disconnected',
    });
  }
}

export const cacheProvider = [
  CacheService,
  {
    provide: CacheAsyncProvider,
    inject: [CacheService],
    useFactory: (cacheService: CacheService) => cacheService.client,
  },
];
