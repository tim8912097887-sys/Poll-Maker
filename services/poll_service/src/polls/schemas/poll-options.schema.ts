import z from 'zod';

export const PollOptionSchema = z.object({
  id: z.uuid(),
  pollId: z.uuid(),
  optionText: z.string().min(1).max(255),
  voteCounts: z.number().min(0),
  createdAt: z.date(),
});

export type PollOptionType = z.infer<typeof PollOptionSchema>;
