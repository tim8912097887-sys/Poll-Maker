import { drizzle, NodePgDatabase } from 'drizzle-orm/node-postgres';
import { Pool } from 'pg';
import { ConfigService } from '@nestjs/config';
import { LoggerService } from '../configs/logging/logger.service';
import { Injectable, OnApplicationShutdown } from '@nestjs/common';

export const DbAsyncProvider = 'DbAsyncProvider';

@Injectable()
class DbService implements OnApplicationShutdown {
  private pool: Pool;
  public db: NodePgDatabase;
  constructor(
    private readonly configService: ConfigService,
    private readonly logger: LoggerService,
  ) {
    const connectionString = this.configService.get<string>('DATABASE_URL');
    this.pool = new Pool({
      connectionString,
      connectionTimeoutMillis: 5000,
      statement_timeout: 10000,
      query_timeout: 12000,
      max: 15,
      min: 2,
      idleTimeoutMillis: 5000,
    });

    // Register event listeners
    this.pool.on('connect', () => {
      logger.info('Database connection established successfully', {
        event: 'database_status',
        status: 'connected',
      });
    });
    this.pool.on('error', (error: any) => {
      logger.error('Database error occurred', {
        event: 'database_status',
        status: 'error',
        error: error.message,
      });
    });
    this.pool.on('acquire', () => {
      logger.info('Database client acquired', {
        event: 'database_status',
        status: 'acquired',
      });
    });
    this.pool.on('release', () => {
      logger.info('Database client released', {
        event: 'database_status',
        status: 'released',
      });
    });
    this.pool.on('remove', () => {
      logger.info('Database client removed', {
        event: 'database_status',
        status: 'removed',
      });
    });
    this.db = drizzle({ client: this.pool });
  }

  async onApplicationShutdown(signal?: string) {
    this.logger.info(`Closing Database pool due to signal: ${signal}`, {
      event: 'database_status',
      status: 'disconnecting',
    });

    await this.pool.end();

    this.logger.info('Database pool closed gracefully', {
      event: 'database_status',
      status: 'disconnected',
    });
  }
}

export const dbProvider = [
  DbService,
  {
    provide: DbAsyncProvider,
    inject: [DbService],
    useFactory: (dbService: DbService) => dbService.db,
  },
];
