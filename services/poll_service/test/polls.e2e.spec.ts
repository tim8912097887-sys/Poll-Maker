import { INestApplication } from '@nestjs/common';
import { AppModule } from 'src/app.module';
import { Test } from '@nestjs/testing';
import { afterAll, beforeAll, beforeEach, describe, expect, it } from 'vitest';
import request from 'supertest';
import { drizzle } from 'drizzle-orm/node-postgres';
import { DbAsyncProvider } from 'src/infrastructure/persistence/db.provider';
import { notBetween, sql } from 'drizzle-orm';
import { seedPolls } from './helpers/seeding/polls';
import { createPollSchema } from './helpers/factory/polls';

describe('Polls API', () => {
  let app: INestApplication;
  let db: ReturnType<typeof drizzle>;

  beforeAll(async () => {
    const moduleRef = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleRef.createNestApplication();

    await app.init();

    db = app.get(DbAsyncProvider);
  });

  beforeEach(async () => {
    await db.execute(sql`
    TRUNCATE TABLE
        polls,
        poll_options
    RESTART IDENTITY CASCADE
    `);
  });

  afterAll(async () => {
    await app.close();
  });

  describe('GET /api/v1/polls', () => {
    it('should return a list of polls', async () => {
      // Arrange
      await seedPolls(db);

      // Act
      const respond = await request(app.getHttpServer()).get('/api/v1/polls');

      // Assert
      expect(respond.status).toBe(200);
      respond.body.data?.polls.forEach((poll) => {
        expect(poll?.isPrivate).toBe(false);
        expect(new Date(poll?.expiredAt as string) > new Date()).toBe(true);
      });
    });
  });

  describe('POST /api/v1/polls', () => {
    it('should create a new poll', async () => {
      // Arrange
      const createPollInfo = createPollSchema();

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(201);
      expect(respond.body.data?.poll?.title).toBe(createPollInfo.title);
    });

    it('should fail to create a new poll if startedAt is not in the future', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        startedAt: new Date().toISOString(),
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('startedAt must be in the future');
    });

    it('should fail to create a new poll if expiredAt is not after startedAt', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        expiredAt: new Date().toISOString(),
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('expiredAt must be after startedAt');
    });

    it('should fail to create a new poll if startedAt is not at least 1 minute in the future', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        startedAt: new Date(Date.now() + 60 * 1000).toISOString(),
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe(
        'startedAt must be at least 1 minute in the future',
      );
    });

    it('should fail to create a new poll if expiredAt is not at least 5 minute after startedAt', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        startedAt: new Date(Date.now() + 2 * 60 * 1000).toISOString(),
        expiredAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('Poll must last at least 5 minutes');
    });

    it('should fail to create a new poll if there are less than 2 options', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        options: [{ optionText: 'Option 1' }],
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('options must be between 2 and 10');
    });

    it('should fail to create a new poll if there are more than 10 options', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        options: Array.from({ length: 11 }, (_, i) => ({
          optionText: `Option ${i + 1}`,
        })),
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('options must be between 2 and 10');
    });

    it('should fail to create a new poll if poll title is less than 1 character', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        title: '',
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('Title must be at least 1 character');
    });

    it('should fail to create a new poll if poll optiontext is less than 1 characters', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        options: [{ optionText: '' }],
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe(
        'Option text must be at least 1 character',
      );
    });

    it('should fail to create a new poll if poll startedAt is invalid', async () => {
      // Arrange
      const createPollInfo = createPollSchema({
        startedAt: 'invalid-date',
      });

      // Act
      const respond = await request(app.getHttpServer())
        .post('/api/v1/polls')
        .send(createPollInfo);

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toBe('Invalid ISO datetime');
    });
  });

  describe('DELETE /api/v1/polls/:id', () => {
    it('should delete a poll', async () => {
      // Arrange
      const poll = await seedPolls(db);
      const id = poll[0].poll.id;

      // Act
      const respond = await request(app.getHttpServer())
        .delete(`/api/v1/polls/${id}`)
        .send();

      // Assert
      expect(respond.status).toBe(200);
      expect(respond.body.data.message).toBe(
        `This action removes a #${id} poll`,
      );
    });

    it('should fail to delete a poll if poll not found', async () => {
      // Arrange
      const id = '00000000-0000-0000-0000-000000000000';

      // Act
      const respond = await request(app.getHttpServer())
        .delete(`/api/v1/polls/${id}`)
        .send();

      // Assert
      expect(respond.status).toBe(404);
      expect(respond.body.message).toBe(`Poll with id ${id} not found`);
    });

    it('should fail to delete a poll if poll id is invalid', async () => {
      // Arrange
      const id = 'invalid-id';

      // Act
      const respond = await request(app.getHttpServer())
        .delete(`/api/v1/polls/${id}`)
        .send();

      // Assert
      expect(respond.status).toBe(400);
      expect(respond.body.message).toContain('uuid');
    });
  });
});
