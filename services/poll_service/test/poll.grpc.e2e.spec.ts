import { INestApplication } from '@nestjs/common';
import {
  ClientGrpc,
  ClientsModule,
  MicroserviceOptions,
  Transport,
} from '@nestjs/microservices';
import { Test } from '@nestjs/testing';
import { sql } from 'drizzle-orm';
import { drizzle } from 'drizzle-orm/node-postgres';
import { AppModule } from 'src/app.module';
import { DbAsyncProvider } from 'src/infrastructure/persistence/db.provider';
import {
  ValidatePollRequest,
  ValidatePollResponse,
  ValidatePollResponse_ValidityReason,
} from 'src/proto/proto/poll';
import { afterAll, afterEach, beforeAll, describe, expect, it } from 'vitest';
import { seedPoll } from './helpers/seeding/polls';
import { RedisClientType } from 'redis';
import { CacheAsyncProvider } from 'src/infrastructure/cache/cache.provider';
import { firstValueFrom, Observable } from 'rxjs';

interface PollService {
  validatePollForVoting(
    request: ValidatePollRequest,
  ): Observable<ValidatePollResponse>;
}

describe('Polls gRPC API', () => {
  let app: INestApplication;
  let db: ReturnType<typeof drizzle>;
  let cache: RedisClientType;
  let client: ClientGrpc;
  let pollService: PollService;

  beforeAll(async () => {
    const grpcOptions = {
      package: 'poll.v1',
      protoPath: './proto/poll.proto',
      url: '0.0.0.0:50051',
    };
    const moduleRef = await Test.createTestingModule({
      imports: [
        AppModule,
        ClientsModule.register([
          {
            name: 'POLL_GRPC_TEST_CLIENT',
            transport: Transport.GRPC,
            options: grpcOptions,
          },
        ]),
      ],
    }).compile();

    app = moduleRef.createNestApplication();

    app.connectMicroservice<MicroserviceOptions>({
      transport: Transport.GRPC,
      options: grpcOptions,
    });

    // Start all microservices attached to the app
    await app.startAllMicroservices();
    await app.init();

    // Obtain gRPC client instance
    client = moduleRef.get<ClientGrpc>('POLL_GRPC_TEST_CLIENT');
    pollService = client.getService<PollService>('PollService');
    // Obtain DB and cache instance
    db = app.get(DbAsyncProvider);
    cache = app.get(CacheAsyncProvider);
  });

  afterEach(async () => {
    await db.execute(sql`
      TRUNCATE TABLE
          polls,
          poll_options
      RESTART IDENTITY CASCADE
      `);

    await cache.flushAll();
  });

  afterAll(async () => {
    await app.close();
  });

  describe('validatePollForVoting', () => {
    describe('Validation', () => {
      it('should return false if request payload is invalid', async () => {
        // Arrange
        const request = { pollId: 'not-a-uuid' };

        // Act
        const respond = await firstValueFrom(
          pollService.validatePollForVoting(request),
        );

        // Assert
        expect(respond.isValid).toBe(false);
        expect(respond.reason).toBe(
          ValidatePollResponse_ValidityReason.POLL_VALIDATION_ERROR,
        );
      });
    });

    describe('Success', () => {
      it('should return true if poll is existing and valid', async () => {
        // Arrange
        const now = new Date();
        const { pollId } = await seedPoll(db, {
          startedAt: new Date(now.getTime() - 60 * 60 * 1000),
          expiredAt: new Date(now.getTime() + 60 * 60 * 1000),
        });

        // Act
        const respond = await firstValueFrom(
          pollService.validatePollForVoting({ pollId }),
        );

        // Assert
        expect(respond.isValid).toBe(true);
        expect(respond.reason).toBe(ValidatePollResponse_ValidityReason.OK);
      });
    });

    describe('Business Logic Errors', () => {
      it('should return false if poll is not existing', async () => {
        // Arrange
        const id = '00000000-0000-0000-0000-000000000000';

        // Act
        const respond = await firstValueFrom(
          pollService.validatePollForVoting({ pollId: id }),
        );

        // Assert
        expect(respond.isValid).toBe(false);
        expect(respond.reason).toBe(
          ValidatePollResponse_ValidityReason.POLL_NOT_FOUND,
        );
      });

      it('should return false if poll is expired', async () => {
        // Arrange
        const now = new Date();
        const { pollId } = await seedPoll(db, {
          startedAt: new Date(now.getTime() - 2 * 60 * 60 * 1000),
          expiredAt: new Date(now.getTime() - 60 * 60 * 1000),
        });

        // Act
        const respond = await firstValueFrom(
          pollService.validatePollForVoting({ pollId }),
        );

        // Assert
        expect(respond.isValid).toBe(false);
        expect(respond.reason).toBe(
          ValidatePollResponse_ValidityReason.POLL_EXPIRED,
        );
      });

      it('should return false if poll is not started yet', async () => {
        // Arrange
        const now = new Date();
        const { pollId } = await seedPoll(db, {
          startedAt: new Date(now.getTime() + 60 * 60 * 1000),
          expiredAt: new Date(now.getTime() + 2 * 60 * 60 * 1000),
        });

        // Act
        const respond = await firstValueFrom(
          pollService.validatePollForVoting({ pollId }),
        );

        // Assert
        expect(respond.isValid).toBe(false);
        expect(respond.reason).toBe(
          ValidatePollResponse_ValidityReason.POLL_NOT_STARTED,
        );
      });
    });
  });
});
