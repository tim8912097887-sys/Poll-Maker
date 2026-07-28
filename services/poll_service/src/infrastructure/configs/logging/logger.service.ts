import { Injectable } from '@nestjs/common';
import { logger } from './logger.config';
@Injectable()
export class LoggerService {
  info(message: string, ...args: any[]) {
    logger.info(message, ...args);
  }

  error(message: string, ...args: any[]) {
    logger.error(message, ...args);
  }

  warn(message: string, ...args: any[]) {
    logger.warn(message, ...args);
  }

  debug(message: string, ...args: any[]) {
    logger.debug(message, ...args);
  }
}
