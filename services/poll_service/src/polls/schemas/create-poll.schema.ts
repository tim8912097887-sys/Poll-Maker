import z from 'zod';
import { PollOptionSchema } from './poll-options.schema';
import { PollSchema } from './polls.schema';

export const CreatePollSchema = PollSchema.omit({
  id: true,
  creatorSession: true,
  createdAt: true,
  startedAt: true,
  expiredAt: true,
})
  .extend({
    startedAt: z.string(),
    expiredAt: z.string(),
    options: z.array(
      PollOptionSchema.omit({
        id: true,
        pollId: true,
        voteCounts: true,
        createdAt: true,
      }),
    ),
  })
  .refine(
    (data) => data.options.length >= 2 && data.options.length <= 10,
    'options must be between 2 and 10',
  )
  .refine(
    ({ startedAt }) => new Date(startedAt) > new Date(Date.now()),
    'startedAt must be in the future',
  )
  .refine(
    ({ startedAt, expiredAt }) => new Date(startedAt) < new Date(expiredAt),
    'expiredAt must be after startedAt',
  );

export type CreatePollType = z.infer<typeof CreatePollSchema>;
