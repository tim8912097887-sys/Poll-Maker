import z from 'zod';

export const VoteCreatedMessageSchema = z
  .object({
    eventId: z.uuid('Invalid UUID for eventId'),
    pollId: z.uuid('Invalid UUID for pollId'),
    optionId: z.uuid('Invalid UUID for optionId'),
    votedAt: z
      .string()
      .regex(
        /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/,
        'Invalid ISO datetime',
      ),
  })
  .refine(({ votedAt }) => !isNaN(Date.parse(votedAt)), 'Invalid ISO datetime');

export type VoteCreatedMessage = z.infer<typeof VoteCreatedMessageSchema>;
