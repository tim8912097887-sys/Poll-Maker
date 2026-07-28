import { Module } from '@nestjs/common';
import { PollsController } from './polls.controller';
import { PollsService } from './polls.service';
import { DBModule } from 'src/infrastructure/persistence/db.module';
import { PollsRepository } from './polls.repository';

@Module({
  imports: [DBModule],
  controllers: [PollsController],
  providers: [PollsService, PollsRepository],
})
export class PollsModule {}
