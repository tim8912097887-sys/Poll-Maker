import { CreatePollResponse } from '../types/polls-response';

export class CreatePollDto {
  static toDto(poll: CreatePollResponse) {
    const pollDto = {
      id: poll.id,
      title: poll.title,
      isPrivate: poll.isPrivate,
      creatorSession: poll.creatorSession,
      startedAt: poll.startedAt,
      expiredAt: poll.expiredAt,
      options: poll.options,
    };
    return pollDto;
  }
}
