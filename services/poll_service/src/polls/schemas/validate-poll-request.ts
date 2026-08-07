import z from 'zod';

export const ValidatePollRequestSchema = z.object({
  pollId: z.uuid('Invalid UUID for pollId'),
});
