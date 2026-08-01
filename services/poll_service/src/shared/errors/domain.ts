import { HttpException, HttpStatus } from '@nestjs/common';

export abstract class DomainError extends HttpException {
  constructor(message: string | object, status: HttpStatus) {
    super(message, status);
  }
}
