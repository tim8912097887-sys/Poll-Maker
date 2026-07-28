import { Module } from '@nestjs/common';
import { dbProvider } from './db.provider';
import { ConfigModule } from '@nestjs/config';
import { LoggerModule } from '../configs/logging/logger.module';

@Module({
  imports: [ConfigModule, LoggerModule],
  providers: [...dbProvider],
  exports: [...dbProvider],
})
export class DBModule {}
