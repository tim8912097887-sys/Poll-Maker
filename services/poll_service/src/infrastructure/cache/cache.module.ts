import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { LoggerModule } from '../configs/logging/logger.module';
import { cacheProvider } from './cache.provider';

@Module({
  imports: [ConfigModule, LoggerModule],
  providers: [...cacheProvider],
  exports: [...cacheProvider],
})
export class CacheModule {}
