import { PollOptionType } from '../schemas/poll-options.schema';
import { PollType } from '../schemas/polls.schema';

export type GetPollsResponse = (Omit<PollType, 'createdAt'> & {
  options: Omit<PollOptionType, 'id' | 'pollId' | 'createdAt'>[];
})[];

export type CreatePollResponse = Omit<PollType, 'createdAt'> & {
  options: Omit<PollOptionType, 'id' | 'pollId' | 'createdAt'>[];
};
