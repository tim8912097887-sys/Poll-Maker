import { GetPollsResponse } from '../types/polls-response';

export class GetPollsDto {
  static toDto(polls: GetPollsResponse) {
    const returnPolls = polls.map((poll) => ({
      id: poll.id,
      title: poll.title,
      isPrivate: poll.isPrivate,
      creatorSession: poll.creatorSession,
      startedAt: poll.startedAt,
      expiredAt: poll.expiredAt,
      options: poll.options,
    }));
    return returnPolls;
  }
}
