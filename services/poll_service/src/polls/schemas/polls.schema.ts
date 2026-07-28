import z from 'zod';

export const PollSchema = z.object({
  id: z.uuid('Invalid UUID'),
  title: z.string().min(1, 'Title must be at least 1 character').max(255),
  isPrivate: z.boolean(),
  creatorSession: z.string(),
  createdAt: z.date(),
  expiredAt: z.date(),
  startedAt: z.date(),
});

export type PollType = z.infer<typeof PollSchema>;
