import { Inject } from '@nestjs/common';
import type { RedisClientType } from 'redis';
import { CacheAsyncProvider } from 'src/infrastructure/cache/cache.provider';
import { PollMeta } from './types/polls-cache-data';
import { createPollKey, createPollOptionsKey } from './utils/cache-key';
import { POLL_DELETED_EVENT_KEY } from './polls.constant';

export class PollsCache {
  constructor(
    @Inject(CacheAsyncProvider)
    private readonly cache: RedisClientType,
  ) {}

  async createPollCache(pollId: string, pollMeta: PollMeta, expiredAt: Date) {
    const key = createPollKey(pollId);
    const result = await this.cache.hSet(key, {
      StartedAt: pollMeta.StartedAt.toISOString(),
      ExpiredAt: pollMeta.ExpiredAt.toISOString(),
      IsPrivate: pollMeta.IsPrivate ? 'true' : 'false',
    });

    const ttlSeconds = Math.floor((expiredAt.getTime() - Date.now()) / 1000);
    await this.cache.expire(key, ttlSeconds);
    return result > 0;
  }

  async createPollOptionsCache(
    pollId: string,
    pollOptions: string[],
    expiredAt: Date,
  ) {
    const key = createPollOptionsKey(pollId);
    const result = await this.cache.sAdd(key, pollOptions);

    const ttlSeconds = Math.floor((expiredAt.getTime() - Date.now()) / 1000);
    await this.cache.expire(key, ttlSeconds);
    return result > 0;
  }

  async deletePollOptionsCache(pollId: string) {
    const key = createPollOptionsKey(pollId);
    const result = await this.cache.del(key);
    return result === 1;
  }

  async deletePollCache(pollId: string) {
    const key = createPollKey(pollId);
    const result = await this.cache.del(key);
    return result === 1;
  }

  async publishPollDelete(pollId: string) {
    await this.cache.publish(
      POLL_DELETED_EVENT_KEY,
      JSON.stringify({ pollId }),
    );
  }
}
