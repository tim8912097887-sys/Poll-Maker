import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class PollNotStarted extends DomainError {
  constructor(pollKey: string) {
    super(`Poll not started: ${pollKey}`, HttpStatus.BAD_REQUEST);
  }
}
