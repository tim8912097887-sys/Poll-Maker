import z from 'zod';
import { PollOptionSchema } from './poll-options.schema';
import { PollSchema } from './polls.schema';

const isoDateTimeRegex =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/;

export const CreatePollSchema = PollSchema.omit({
  id: true,
  creatorSession: true,
  createdAt: true,
  startedAt: true,
  expiredAt: true,
})
  .extend({
    startedAt: z.string().regex(isoDateTimeRegex, 'Invalid ISO datetime'),
    expiredAt: z.string().regex(isoDateTimeRegex, 'Invalid ISO datetime'),
    options: z.array(
      PollOptionSchema.omit({
        id: true,
        pollId: true,
        voteCounts: true,
        createdAt: true,
      }),
    ),
  })
  //  At least 2 options
  .refine(
    (data) => data.options.length >= 2 && data.options.length <= 10,
    'options must be between 2 and 10',
  )
  // StartedAt must be in the future
  .refine(
    ({ startedAt }) => new Date(startedAt) > new Date(Date.now()),
    'startedAt must be in the future',
  )
  // ExpiredAt must be after startedAt
  .refine(
    ({ startedAt, expiredAt }) => new Date(startedAt) < new Date(expiredAt),
    'expiredAt must be after startedAt',
  )
  // StartedAt must be at least 1 minute in the future
  .refine(
    ({ startedAt }) => new Date(startedAt).getTime() > Date.now() + 60_000,
    'startedAt must be at least 1 minute in the future',
  )
  // Poll must last at least 5 minutes
  .refine(
    ({ startedAt, expiredAt }) =>
      new Date(expiredAt).getTime() - new Date(startedAt).getTime() >=
      5 * 60_000,
    'Poll must last at least 5 minutes',
  );

export type CreatePollType = z.infer<typeof CreatePollSchema>;
