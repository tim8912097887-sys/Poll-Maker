import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class CacheNotFound extends DomainError {
  constructor(cacheKey: string) {
    super(`Cache not found: ${cacheKey}`, HttpStatus.NOT_FOUND);
  }
}
