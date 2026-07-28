import { randomUUID } from 'crypto';
import { drizzle } from 'drizzle-orm/node-postgres';
import { pollOptions } from 'src/infrastructure/persistence/schemas/poll_options';
import { polls } from 'src/infrastructure/persistence/schemas/polls';

export async function seedPolls(db: ReturnType<typeof drizzle>) {
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
    await db.insert(polls).values(item.poll);

    await db.insert(pollOptions).values(
      item.options.map((text) => ({
        id: randomUUID(),
        pollId: item.poll.id,
        optionText: text,
      })),
    );
  }

  return data;
}
