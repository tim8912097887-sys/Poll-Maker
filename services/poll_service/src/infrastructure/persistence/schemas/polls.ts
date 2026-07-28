import {
  boolean,
  pgTable,
  timestamp,
  uuid,
  varchar,
} from 'drizzle-orm/pg-core';

export const polls = pgTable('polls', {
  id: uuid('id').primaryKey().defaultRandom(),
  title: varchar('title', { length: 255 }).notNull(),
  isPrivate: boolean('is_private').notNull(),
  creatorSession: varchar('creator_session', { length: 255 }).notNull(),
  createdAt: timestamp('created_at', { withTimezone: true })
    .defaultNow()
    .notNull(),
  startedAt: timestamp('started_at', { withTimezone: true }).notNull(),
  expiredAt: timestamp('expired_at', { withTimezone: true }).notNull(),
});
