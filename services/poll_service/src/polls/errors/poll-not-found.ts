import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class PollNotFound extends DomainError {
  constructor(pollKey: string) {
    super(`Poll not found: ${pollKey}`, HttpStatus.NOT_FOUND);
  }
}
