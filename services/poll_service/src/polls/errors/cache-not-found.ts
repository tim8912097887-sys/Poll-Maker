import { DomainError } from 'src/shared/errors/domain';

export class CacheNotFound extends DomainError {
  readonly status = 404;
  constructor(cacheKey: string) {
    super(`Cache not found: ${cacheKey}`);
  }
}
