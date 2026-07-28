import { drizzle } from 'drizzle-orm/node-postgres';
import { Pool } from 'pg';
import { ConfigService } from '@nestjs/config';
import { LoggerService } from '../configs/logging/logger.service';

export const DbAsyncProvider = 'DbAsyncProvider';

export const dbProvider = [
  {
    provide: DbAsyncProvider,
    inject: [ConfigService, LoggerService],
    useFactory: (configService: ConfigService, logger: LoggerService) => {
      const connectionString = configService.get<string>('DATABASE_URL');
      const pool = new Pool({
        connectionString,
        connectionTimeoutMillis: 5000,
        statement_timeout: 10000,
        query_timeout: 12000,
        max: 15,
        min: 2,
        idleTimeoutMillis: 5000,
      });

      // Register event listeners
      pool.on('connect', () => {
        logger.info('Database connection established successfully', {
          event: 'database_status',
          status: 'connected',
        });
      });
      pool.on('error', (error: any) => {
        logger.error('Database error occurred', {
          event: 'database_status',
          status: 'error',
          error: error.message,
        });
      });
      pool.on('acquire', () => {
        logger.info('Database client acquired', {
          event: 'database_status',
          status: 'acquired',
        });
      });
      pool.on('release', () => {
        logger.info('Database client released', {
          event: 'database_status',
          status: 'released',
        });
      });
      pool.on('remove', () => {
        logger.info('Database client removed', {
          event: 'database_status',
          status: 'removed',
        });
      });
      return drizzle({ client: pool });
    },
  },
];
