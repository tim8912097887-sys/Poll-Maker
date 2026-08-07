import { pgTable, timestamp, uuid } from 'drizzle-orm/pg-core';

export const inboxEvents = pgTable('inbox_events', {
  eventId: uuid('event_id').primaryKey(),
  processedAt: timestamp('processed_at', { withTimezone: true })
    .defaultNow()
    .notNull(),
});
