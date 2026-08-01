import {
  ExceptionFilter,
  Catch,
  ArgumentsHost,
  HttpException,
  HttpStatus,
} from '@nestjs/common';
import { Request, Response } from 'express';
import { errorResponse } from '../response/error';
import { logger } from 'src/infrastructure/configs/logging/logger.config';
import { DomainError } from '../errors/domain';

const STATUS_CODE_MAPPING: Record<number, string> = {
  400: 'Bad Request',
  401: 'Unauthorized',
  403: 'Forbidden',
  404: 'Not Found',
  409: 'Conflict',
  500: 'Internal Server Error',
};

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    const status =
      exception instanceof HttpException
        ? exception.getStatus()
        : HttpStatus.INTERNAL_SERVER_ERROR;

    const message =
      exception instanceof HttpException
        ? exception.getResponse()
        : 'Internal Server Error';

    if (!(exception instanceof HttpException)) {
      logger.error(
        {
          event: 'server_error',
          error: exception instanceof Error ? exception.stack : message,
        },
        `Error occurred on ${request.method} ${request.url}`,
      );
    }

    // Safely extract the message string/object
    const errorDetails =
      typeof message === 'object' && message !== null && 'message' in message
        ? ((message as Record<string, any>).message as string)
        : (message as string);
    response
      .status(status)
      .json(errorResponse(STATUS_CODE_MAPPING[status], errorDetails));
  }
}
