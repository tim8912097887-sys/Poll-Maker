import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { PollsModule } from './polls/polls.module';
import { ConfigModule } from '@nestjs/config';
import env from './infrastructure/configs/env';
import { DBModule } from './infrastructure/persistence/db.module';
import { LoggerModule } from './infrastructure/configs/logging/logger.module';
import { CacheModule } from './infrastructure/cache/cache.module';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { CLIENT_ID, GROUP_ID } from './constants';

@Module({
  imports: [
    ConfigModule.forRoot({
      load: [env],
    }),
    ClientsModule.register([
      {
        name: 'KAFKA_CLIENT',
        transport: Transport.KAFKA,
        options: {
          client: {
            clientId: CLIENT_ID,
            brokers: [process.env.KAFKA_BROKERS ?? 'localhost:9092'],
          },
          consumer: {
            groupId: GROUP_ID,
          },
        },
      },
    ]),
    LoggerModule,
    DBModule,
    CacheModule,
    PollsModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
