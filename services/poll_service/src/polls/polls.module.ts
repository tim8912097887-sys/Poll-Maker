import { Module } from '@nestjs/common';
import { PollsController } from './polls.controller';
import { PollsService } from './polls.service';
import { DBModule } from 'src/infrastructure/persistence/db.module';
import { PollsRepository } from './polls.repository';
import { CacheModule } from 'src/infrastructure/cache/cache.module';
import { PollsCache } from './polls.cache';

@Module({
  imports: [DBModule, CacheModule],
  controllers: [PollsController],
  providers: [PollsService, PollsRepository, PollsCache],
})
export class PollsModule {}
