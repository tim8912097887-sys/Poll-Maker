import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { PollsModule } from './polls/polls.module';
import { ConfigModule } from '@nestjs/config';
import env from './infrastructure/configs/env';
import { LoggerService } from './infrastructure/configs/logging/logger.service';
import { DBModule } from './infrastructure/persistence/db.module';
import { LoggerModule } from './infrastructure/configs/logging/logger.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      load: [env],
    }),
    LoggerModule,
    DBModule,
    PollsModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
