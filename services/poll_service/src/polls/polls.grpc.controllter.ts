import { Controller } from '@nestjs/common';
import {
  PollServiceController,
  PollServiceControllerMethods,
  type ValidatePollRequest,
  ValidatePollResponse,
  ValidatePollResponse_ValidityReason,
} from 'src/proto/proto/poll';
import { PollsService } from './polls.service';
import { GrpcMethod } from '@nestjs/microservices';
import { PollNotFound } from './errors/poll-not-found';
import { PollExpired } from './errors/poll-expired';
import { PollNotStarted } from './errors/poll-not-started';
import { toTimestamp } from './utils/time-convertion';
import { logger } from 'src/infrastructure/configs/logging/logger.config';

@Controller()
@PollServiceControllerMethods()
export class PollGrpcController implements PollServiceController {
  constructor(private readonly pollsService: PollsService) {}

  @GrpcMethod('PollService', 'ValidatePollForVoting')
  async validatePollForVoting(
    request: ValidatePollRequest,
  ): Promise<ValidatePollResponse> {
    try {
      const poll = await this.pollsService.validatePollForVoting(request);
      return {
        isValid: true,
        reason: ValidatePollResponse_ValidityReason.OK,
        expiredAt: toTimestamp(poll.expiredAt),
      };
    } catch (error) {
      if (error instanceof PollNotFound) {
        logger.error({
          event: 'poll_not_found',
          pollId: request.pollId,
        });
        return {
          isValid: false,
          reason: ValidatePollResponse_ValidityReason.POLL_NOT_FOUND,
          expiredAt: toTimestamp(new Date()),
        };
      } else if (error instanceof PollExpired) {
        logger.error({
          event: 'poll_expired',
          pollId: request.pollId,
        });
        return {
          isValid: false,
          reason: ValidatePollResponse_ValidityReason.POLL_EXPIRED,
          expiredAt: toTimestamp(new Date()),
        };
      } else if (error instanceof PollNotStarted) {
        logger.error({
          event: 'poll_not_started',
          pollId: request.pollId,
        });
        return {
          isValid: false,
          reason: ValidatePollResponse_ValidityReason.POLL_NOT_STARTED,
          expiredAt: toTimestamp(new Date()),
        };
      } else {
        logger.error({
          event: 'unknown_error',
          pollId: request.pollId,
        });
        return {
          isValid: false,
          reason: ValidatePollResponse_ValidityReason.UNRECOGNIZED,
          expiredAt: toTimestamp(new Date()),
        };
      }
    }
  }
}
