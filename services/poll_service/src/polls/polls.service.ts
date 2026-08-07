import { Injectable, NotFoundException } from '@nestjs/common';
import { CreatePollResponse, GetPollsResponse } from './types/polls-response';
import { CreatePollType } from './schemas/create-poll.schema';
import { PollsRepository } from './polls.repository';
import { PollsCache } from './polls.cache';
import { PollMeta } from './types/polls-cache-data';
import { CacheNotFound } from './errors/cache-not-found';
import { ValidatePollRequest } from 'src/proto/proto/poll';
import { PollNotFound } from './errors/poll-not-found';
import { PollExpired } from './errors/poll-expired';
import { PollNotStarted } from './errors/poll-not-started';
import { PollOptionNotFound } from './errors/poll-option-not-found';

@Injectable()
export class PollsService {
  constructor(
    private readonly pollsRepository: PollsRepository,
    private readonly pollsCache: PollsCache,
  ) {}
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
    // Create in db
    const createdPoll = await this.pollsRepository.createPoll(
      pollId,
      creatorSession,
      createPollSchema,
    );

    const pollMeta: PollMeta = {
      StartedAt: createdPoll.startedAt,
      ExpiredAt: createdPoll.expiredAt,
      IsPrivate: createdPoll.isPrivate,
    };
    // Create in cache
    const createdPollCacheResult = await this.pollsCache.createPollCache(
      createdPoll.id,
      pollMeta,
      createdPoll.expiredAt,
    );
    const createdPollOptionsCacheResult =
      await this.pollsCache.createPollOptionsCache(
        createdPoll.id,
        createdPoll.options,
        createdPoll.expiredAt,
      );
    const isCacheCreated =
      createdPollCacheResult && createdPollOptionsCacheResult;
    if (!isCacheCreated) {
      throw new Error('Failed to create poll in cache');
    }

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
    // Delete from db
    const result = await this.pollsRepository.deletePoll(id);

    if (result.rowCount === 0) {
      throw new NotFoundException(`Poll with id ${id} not found`);
    }

    // Delete from cache
    const pollCacheDeletedResult = await this.pollsCache.deletePollCache(id);
    const pollOptionsCacheDeletedResult =
      await this.pollsCache.deletePollOptionsCache(id);
    const isCacheDeleted =
      pollCacheDeletedResult && pollOptionsCacheDeletedResult;
    if (!isCacheDeleted) {
      throw new CacheNotFound(id);
    }

    // Publish delete event
    await this.pollsCache.publishPollDelete(id);

    return `This action removes a #${id} poll`;
  }

  async validatePollForVoting(validatePollRequest: ValidatePollRequest) {
    const poll = await this.checkPollStatus(validatePollRequest.pollId);
    return poll;
  }

  async updateVoteCount(eventId: string, pollId: string, optionId: string) {
    const poll = await this.checkPollStatus(pollId);
    if (!poll.options.includes(optionId)) {
      throw new PollOptionNotFound(pollId, optionId);
    }
    await this.pollsRepository.updatePollOption(eventId, pollId, optionId);
  }

  private async checkPollStatus(pollId: string) {
    const poll = await this.pollsRepository.findPollById(pollId);
    if (!poll) {
      throw new PollNotFound(pollId);
    }
    if (poll.expiredAt < new Date()) {
      throw new PollExpired(pollId);
    }
    if (poll.startedAt > new Date()) {
      throw new PollNotStarted(pollId);
    }
    return poll;
  }
}
