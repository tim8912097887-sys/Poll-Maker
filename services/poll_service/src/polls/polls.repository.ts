import { Inject } from '@nestjs/common';
import { and, eq, gt, sql } from 'drizzle-orm';
import { NodePgDatabase } from 'drizzle-orm/node-postgres';
import { DbAsyncProvider } from 'src/infrastructure/persistence/db.provider';
import { polls } from 'src/infrastructure/persistence/schemas/polls';
import { CreatePollType } from './schemas/create-poll.schema';
import { pollOptions } from 'src/infrastructure/persistence/schemas/poll_options';
import { inboxEvents } from 'src/infrastructure/persistence/schemas/inbox_events';

export class PollsRepository {
  constructor(
    @Inject(DbAsyncProvider)
    private readonly db: NodePgDatabase,
  ) {}

  async findPublicPolls() {
    return this.db
      .select()
      .from(polls)
      .where(and(eq(polls.isPrivate, false), gt(polls.expiredAt, new Date())));
  }

  async findPollById(id: string) {
    const [poll] = await this.db
      .select({
        id: polls.id,
        title: polls.title,
        isPrivate: polls.isPrivate,
        creatorSession: polls.creatorSession,
        createdAt: polls.createdAt,
        startedAt: polls.startedAt,
        expiredAt: polls.expiredAt,
        options: sql<string[]>`json_agg(${pollOptions.id})`,
      })
      .from(polls)
      .innerJoin(pollOptions, eq(polls.id, pollOptions.pollId))
      .where(eq(polls.id, id))
      .groupBy(
        polls.id,
        polls.title,
        polls.isPrivate,
        polls.creatorSession,
        polls.createdAt,
        polls.startedAt,
        polls.expiredAt,
      );

    return poll;
  }

  async createPoll(
    id: string,
    creatorSession: string,
    createPollSchema: CreatePollType,
  ) {
    await this.db.transaction(async (tx) => {
      const { options, ...rest } = createPollSchema;
      // Insert poll
      await tx.insert(polls).values({
        ...rest,
        id,
        creatorSession,
        startedAt: new Date(rest.startedAt),
        expiredAt: new Date(rest.expiredAt),
      });

      // Insert poll options
      const insertedOptions = options.map((option) => ({
        ...option,
        id: crypto.randomUUID(),
        pollId: id,
      }));

      await tx.insert(pollOptions).values(insertedOptions);
    });

    // Get created poll
    const createdPoll = await this.findPollById(id);

    return createdPoll;
  }

  async deletePoll(id: string) {
    return this.db.delete(polls).where(eq(polls.id, id));
  }

  async updatePollOption(eventId: string, pollId: string, optionId: string) {
    const voteCounts = await this.db.transaction(async (tx) => {
      // Prevent race condition by directly updating the row
      const [{ voteCounts }] = await tx
        .update(pollOptions)
        .set({ voteCounts: sql<number>`vote_counts + 1` })
        .where(
          and(eq(pollOptions.id, optionId), eq(pollOptions.pollId, pollId)),
        )
        .returning({ id: pollOptions.id, voteCounts: pollOptions.voteCounts });

      // Create inbox record
      await tx.insert(inboxEvents).values({ eventId });

      return voteCounts;
    });

    return voteCounts;
  }
}
