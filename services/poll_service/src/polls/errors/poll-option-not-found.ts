import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class PollOptionNotFound extends DomainError {
  constructor(pollKey: string, optionKey: string) {
    super(
      `Poll option not found: ${pollKey}, ${optionKey}`,
      HttpStatus.NOT_FOUND,
    );
  }
}
