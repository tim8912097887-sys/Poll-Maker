import { HttpStatus } from '@nestjs/common';
import { DomainError } from 'src/shared/errors/domain';

export class PollValidateFail extends DomainError {
  constructor(message: string) {
    super(`Poll validate fail: ${message}`, HttpStatus.BAD_REQUEST);
  }
}
