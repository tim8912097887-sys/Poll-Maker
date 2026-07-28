import { Injectable, NotFoundException } from '@nestjs/common';
import { CreatePollResponse, GetPollsResponse } from './types/polls-response';
import { CreatePollType } from './schemas/create-poll.schema';
import { PollsRepository } from './polls.repository';

@Injectable()
export class PollsService {
  constructor(private readonly pollsRepository: PollsRepository) {}
  async getPublicPolls(): Promise<GetPollsResponse> {
    const polls = await this.pollsRepository.findPublicPolls();
    const pollResponse = polls.map((poll) => ({
      id: poll.id,
      title: poll.title,
      isPrivate: poll.isPrivate,
      creatorSession: poll.creatorSession,
      startedAt: poll.startedAt,
      expiredAt: poll.expiredAt,
    }));
    return pollResponse;
  }

  async createPoll(
    createPollSchema: CreatePollType,
  ): Promise<CreatePollResponse> {
    const pollId = crypto.randomUUID();
    const creatorSession = crypto.randomUUID();
    const createdPoll = await this.pollsRepository.createPoll(
      pollId,
      creatorSession,
      createPollSchema,
    );

    const createdPollResponse = {
      id: createdPoll.id,
      title: createdPoll.title,
      isPrivate: createdPoll.isPrivate,
      creatorSession: createdPoll.creatorSession,
      startedAt: createdPoll.startedAt,
      expiredAt: createdPoll.expiredAt,
      options: createdPoll.options.map((option) => ({
        optionText: option,
      })),
    };
    return createdPollResponse;
  }

  async deletePoll(id: string): Promise<string> {
    const result = await this.pollsRepository.deletePoll(id);

    if (result.rowCount === 0) {
      throw new NotFoundException(`Poll with id ${id} not found`);
    }

    return `This action removes a #${id} poll`;
  }
}
