import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class PollExpired extends DomainError {
  constructor(pollKey: string) {
    super(`Poll expired: ${pollKey}`, HttpStatus.BAD_REQUEST);
  }
}
