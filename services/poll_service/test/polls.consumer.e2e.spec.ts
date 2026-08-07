import { INestApplication } from '@nestjs/common';
import {
  ClientKafka,
  ClientsModule,
  MicroserviceOptions,
  Transport,
} from '@nestjs/microservices';
import { Test } from '@nestjs/testing';
import { sql } from 'drizzle-orm';
import { drizzle } from 'drizzle-orm/node-postgres';
import { RedisClientType } from 'redis';
import { AppModule } from 'src/app.module';
import { CacheAsyncProvider } from 'src/infrastructure/cache/cache.provider';
import { logger } from 'src/infrastructure/configs/logging/logger.config';
import { DbAsyncProvider } from 'src/infrastructure/persistence/db.provider';
import { PollExpired } from 'src/polls/errors/poll-expired';
import { PollNotFound } from 'src/polls/errors/poll-not-found';
import { PollNotStarted } from 'src/polls/errors/poll-not-started';
import { PollOptionNotFound } from 'src/polls/errors/poll-option-not-found';
import { VOTE_CREATED_TOPIC } from 'src/polls/polls.constant';
import { seedPoll } from './helpers/seeding/polls';
import {
  afterAll,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vitest,
} from 'vitest';

describe('Polls Consumer API', () => {
  let app: INestApplication;
  let db: ReturnType<typeof drizzle>;
  let cache: RedisClientType;
  let client: ClientKafka;

  beforeAll(async () => {
    const kafkaOptions = {
      client: {
        clientId: `poll-service-test-client-${crypto.randomUUID()}`,
        brokers: ['localhost:29092'],
        retry: {
          retries: 10,
          initialRetryTime: 3000,
        },
      },
      consumer: {
        // Generate random group id prevent collisions
        groupId: `poll-service-consumer-test-${crypto.randomUUID()}`,
        allowAutoTopicCreation: true,
      },
    };
    const moduleRef = await Test.createTestingModule({
      imports: [
        AppModule,
        ClientsModule.register([
          {
            name: 'POLL_CONSUMER_TEST_CLIENT',
            transport: Transport.KAFKA,
            options: kafkaOptions,
          },
        ]),
      ],
    }).compile();

    app = moduleRef.createNestApplication();

    app.connectMicroservice<MicroserviceOptions>({
      transport: Transport.KAFKA,
      options: kafkaOptions,
    });

    // Start all microservices attached to the app
    await app.startAllMicroservices();
    // Wait for the consumer to be ready
    await new Promise((resolve) => setTimeout(resolve, 3000));
    await app.init();

    // Obtain kafka client instance
    client = moduleRef.get<ClientKafka>('POLL_CONSUMER_TEST_CLIENT');
    await client.connect();
    // Obtain DB and cache instance
    db = app.get(DbAsyncProvider);
    cache = app.get(CacheAsyncProvider);
  });

  beforeEach(async () => {
    await db.execute(sql`
      TRUNCATE TABLE
          polls,
          poll_options,
          inbox_events
      RESTART IDENTITY CASCADE
      `);

    await cache.flushAll();
  });

  afterAll(async () => {
    await app.close();
  });

  describe('Vote Created Event', () => {
    const sendVoteCreatedEvent = async (message: unknown) => {
      await new Promise<void>((resolve, reject) => {
        client.emit(VOTE_CREATED_TOPIC, message).subscribe({
          next: () => resolve(),
          error: (err) => reject(err),
          complete: () => resolve(),
        });
      });
    };

    const getInboxEventsCount = async () => {
      const result = await db.execute(sql`SELECT * FROM inbox_events`);
      return result.rows.length;
    };

    const getVoteCount = async (optionId: string) => {
      const result = await db.execute(
        sql`SELECT vote_counts FROM poll_options WHERE id = ${optionId}`,
      );
      return result.rows[0]?.vote_counts ?? 0;
    };

    describe('Validation', () => {
      it('should fail to update vote counts if eventId is invalid', async () => {
        // Arrange
        const voteCreatedMessage = {
          eventId: 'invalid-uuid',
          pollId: crypto.randomUUID(),
          optionId: crypto.randomUUID(),
          votedAt: new Date().toISOString(),
        };

        // Act
        await sendVoteCreatedEvent(voteCreatedMessage);

        // Assert
        await new Promise((resolve) => setTimeout(resolve, 500));
        expect(await getInboxEventsCount()).toBe(0);
      });

      it('should fail to update vote counts if pollId is invalid', async () => {
        // Arrange
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId: 'invalid-uuid',
          optionId: crypto.randomUUID(),
          votedAt: new Date().toISOString(),
        };

        // Act
        await sendVoteCreatedEvent(voteCreatedMessage);

        // Assert
        await new Promise((resolve) => setTimeout(resolve, 500));
        expect(await getInboxEventsCount()).toBe(0);
      });

      it('should fail to update vote counts if optionId is invalid', async () => {
        // Arrange
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId: crypto.randomUUID(),
          optionId: 'invalid-uuid',
          votedAt: new Date().toISOString(),
        };

        // Act
        await sendVoteCreatedEvent(voteCreatedMessage);

        // Assert
        await new Promise((resolve) => setTimeout(resolve, 500));
        expect(await getInboxEventsCount()).toBe(0);
      });

      it('should fail to update vote counts if votedAt is invalid', async () => {
        // Arrange
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId: crypto.randomUUID(),
          optionId: crypto.randomUUID(),
          votedAt: 'not-a-valid-date',
        };

        // Act
        await sendVoteCreatedEvent(voteCreatedMessage);

        // Assert
        await new Promise((resolve) => setTimeout(resolve, 500));
        expect(await getInboxEventsCount()).toBe(0);
      });
    });

    describe('Business Logic Errors', () => {
      it('should fail to update vote counts if poll not found', async () => {
        // Arrange
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId: crypto.randomUUID(),
          optionId: crypto.randomUUID(),
          votedAt: new Date().toISOString(),
        };
        const errorSpy = vitest.spyOn(logger, 'error');

        // Act
        await sendVoteCreatedEvent(voteCreatedMessage);

        // Assert
        expect(await getInboxEventsCount()).toBe(0);
        await vitest.waitFor(
          () => {
            expect(errorSpy).toHaveBeenCalledWith({
              event: 'vote_created_error',
              error: new PollNotFound(voteCreatedMessage.pollId),
            });
          },
          { timeout: 5000, interval: 100 },
        );
      });

      it('should fail to update vote counts if poll option is not found', async () => {
        const { pollId } = await seedPoll(db, {
          startedAt: new Date(Date.now() - 60 * 60 * 1000),
          expiredAt: new Date(Date.now() + 60 * 60 * 1000),
        });
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId,
          optionId: crypto.randomUUID(),
          votedAt: new Date().toISOString(),
        };
        const errorSpy = vitest.spyOn(logger, 'error');

        await sendVoteCreatedEvent(voteCreatedMessage);

        expect(await getInboxEventsCount()).toBe(0);
        await vitest.waitFor(
          () => {
            expect(errorSpy).toHaveBeenCalledWith({
              event: 'vote_created_error',
              error: new PollOptionNotFound(
                voteCreatedMessage.pollId,
                voteCreatedMessage.optionId,
              ),
            });
          },
          { timeout: 5000, interval: 100 },
        );
      });

      it('should fail to update vote counts if poll has already expired', async () => {
        const { pollId, optionId } = await seedPoll(db, {
          startedAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
          expiredAt: new Date(Date.now() - 60 * 60 * 1000),
        });
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId,
          optionId,
          votedAt: new Date().toISOString(),
        };
        const errorSpy = vitest.spyOn(logger, 'error');

        await sendVoteCreatedEvent(voteCreatedMessage);

        expect(await getInboxEventsCount()).toBe(0);
        await vitest.waitFor(
          () => {
            expect(errorSpy).toHaveBeenCalledWith({
              event: 'vote_created_error',
              error: new PollExpired(voteCreatedMessage.pollId),
            });
          },
          { timeout: 5000, interval: 100 },
        );
      });

      it('should fail to update vote counts if poll has not started yet', async () => {
        const { pollId, optionId } = await seedPoll(db, {
          startedAt: new Date(Date.now() + 60 * 60 * 1000),
          expiredAt: new Date(Date.now() + 2 * 60 * 60 * 1000),
        });
        const voteCreatedMessage = {
          eventId: crypto.randomUUID(),
          pollId,
          optionId,
          votedAt: new Date().toISOString(),
        };
        const errorSpy = vitest.spyOn(logger, 'error');

        await sendVoteCreatedEvent(voteCreatedMessage);

        expect(await getInboxEventsCount()).toBe(0);
        await vitest.waitFor(
          () => {
            expect(errorSpy).toHaveBeenCalledWith({
              event: 'vote_created_error',
              error: new PollNotStarted(voteCreatedMessage.pollId),
            });
          },
          { timeout: 5000, interval: 100 },
        );
      });

      it('should ignore duplicate vote events by eventId', async () => {
        const { pollId, optionId } = await seedPoll(db, {
          startedAt: new Date(Date.now() - 60 * 60 * 1000),
          expiredAt: new Date(Date.now() + 60 * 60 * 1000),
        });
        const eventId = crypto.randomUUID();
        const voteCreatedMessage = {
          eventId,
          pollId,
          optionId,
          votedAt: new Date().toISOString(),
        };

        await sendVoteCreatedEvent(voteCreatedMessage);
        await new Promise((resolve) => setTimeout(resolve, 500));
        await sendVoteCreatedEvent(voteCreatedMessage);

        expect(await getInboxEventsCount()).toBe(1);
        expect(await getVoteCount(optionId)).toBe(1);
      });
    });
  });
});
