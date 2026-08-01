import { randomUUID } from 'crypto';
import { drizzle } from 'drizzle-orm/node-postgres';
import { RedisClientType } from 'redis';
import { pollOptions } from 'src/infrastructure/persistence/schemas/poll_options';
import { polls } from 'src/infrastructure/persistence/schemas/polls';
import { createPollKey, createPollOptionsKey } from 'src/polls/utils/cache-key';

export async function seedPolls(
  db: ReturnType<typeof drizzle>,
  cache: RedisClientType,
) {
  const now = new Date();

  const data = [
    {
      poll: {
        id: randomUUID(),
        title: 'Favorite Programming Language',
        isPrivate: false,
        creatorSession: 'test' + randomUUID(),
        startedAt: new Date(now.getTime() - 24 * 60 * 60 * 1000),
        expiredAt: new Date(now.getTime() - 12 * 60 * 60 * 1000),
      },
      options: ['TypeScript', 'Go', 'Rust', 'Python'],
    },
    {
      poll: {
        id: randomUUID(),
        title: 'Best Backend Framework',
        isPrivate: false,
        creatorSession: 'test' + randomUUID(),
        startedAt: new Date(now.getTime() - 48 * 60 * 60 * 1000),
        expiredAt: new Date(now.getTime() - 12 * 60 * 60 * 1000),
      },
      options: ['NestJS', 'Express', 'Fastify', 'Gin'],
    },
    {
      poll: {
        id: randomUUID(),
        title: 'Favorite Database',
        isPrivate: false,
        creatorSession: 'test' + randomUUID(),
        startedAt: new Date(now.getTime() + 12 * 60 * 60 * 1000),
        expiredAt: new Date(now.getTime() + 72 * 60 * 60 * 1000),
      },
      options: ['PostgreSQL', 'MySQL', 'MongoDB', 'SQLite'],
    },
    {
      poll: {
        id: randomUUID(),
        title: 'Favorite Cloud Provider',
        isPrivate: false,
        creatorSession: 'test' + randomUUID(),
        startedAt: new Date(now.getTime() + 24 * 60 * 60 * 1000),
        expiredAt: new Date(now.getTime() + 96 * 60 * 60 * 1000),
      },
      options: ['AWS', 'Azure', 'Google Cloud', 'DigitalOcean'],
    },
    {
      poll: {
        id: randomUUID(),
        title: 'Favorite Frontend Framework',
        isPrivate: false,
        creatorSession: 'test' + randomUUID(),
        startedAt: new Date(now.getTime() + 48 * 60 * 60 * 1000),
        expiredAt: new Date(now.getTime() + 120 * 60 * 60 * 1000),
      },
      options: ['React', 'Vue', 'Angular', 'Svelte'],
    },
  ];

  for (const item of data) {
    // Insert poll and options into the database
    await db.insert(polls).values(item.poll);

    await db.insert(pollOptions).values(
      item.options.map((text) => ({
        id: randomUUID(),
        pollId: item.poll.id,
        optionText: text,
      })),
    );

    // Insert poll and options into the cache
    const pollKey = createPollKey(item.poll.id);
    await cache.hSet(pollKey, {
      StartedAt: item.poll.startedAt.toISOString(),
      ExpiredAt: item.poll.expiredAt.toISOString(),
      IsPrivate: item.poll.isPrivate ? 'true' : 'false',
    });

    await cache.expire(
      pollKey,
      Math.floor((item.poll.expiredAt.getTime() - Date.now()) / 1000),
    );

    const pollOptionsKey = createPollOptionsKey(item.poll.id);
    await cache.sAdd(pollOptionsKey, item.options);

    await cache.expire(
      pollOptionsKey,
      Math.floor((item.poll.expiredAt.getTime() - Date.now()) / 1000),
    );
  }

  return data;
}
