import z from 'zod';

export const PollSchema = z.object({
  id: z.uuid(),
  title: z.string().min(1).max(255),
  isPrivate: z.boolean(),
  creatorSession: z.string(),
  createdAt: z.date(),
  expiredAt: z.date(),
  startedAt: z.date(),
});

export type PollType = z.infer<typeof PollSchema>;
