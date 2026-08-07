import {
  Body,
  Controller,
  Delete,
  Get,
  Param,
  ParseUUIDPipe,
  Post,
} from '@nestjs/common';
import { PollsService } from './polls.service';
import {
  CreatePollSchema,
  type CreatePollType,
} from './schemas/create-poll.schema';
import { ZodValidationPipe } from 'src/shared/pipes/validation.pipe';
import { GetPollsDto } from './dto/get-polls.dto';
import { CreatePollDto } from './dto/create-poll.dto';
import { successResponse } from 'src/shared/response/success';
import { ServerSuccess } from 'src/shared/response/types';
import {
  Ctx,
  EventPattern,
  KafkaContext,
  Payload,
} from '@nestjs/microservices';
import type { VoteCreatedMessage } from './types/polls-event';
import { logger } from 'src/infrastructure/configs/logging/logger.config';

@Controller('/api/v1/polls')
export class PollsController {
  constructor(private readonly pollsService: PollsService) {}

  @Get()
  async getPublicPolls(): Promise<
    ServerSuccess<{ polls: GetPollsDto; message: string }>
  > {
    const polls = await this.pollsService.getPublicPolls();
    const data = {
      polls: GetPollsDto.toDto(polls),
      message: 'Successfully fetched public polls',
    };
    return successResponse(data);
  }

  @Post()
  async createPoll(
    @Body(new ZodValidationPipe(CreatePollSchema))
    createPollSchema: CreatePollType,
  ): Promise<ServerSuccess<{ poll: CreatePollDto; message: string }>> {
    const poll = await this.pollsService.createPoll(createPollSchema);
    const data = {
      poll: CreatePollDto.toDto(poll),
      message: 'Successfully created poll',
    };
    return successResponse(data);
  }

  @Delete(':id')
  async deletePoll(
    @Param('id', new ParseUUIDPipe()) id: string,
  ): Promise<ServerSuccess<{ message: string }>> {
    const result = await this.pollsService.deletePoll(id);
    const data = { message: result };
    return successResponse(data);
  }

  @EventPattern('vote.created')
  async handleVoteCreated(
    @Payload() message: VoteCreatedMessage,
    @Ctx() context: KafkaContext,
  ) {
    try {
      await this.pollsService.updateVoteCount(
        message.eventId,
        message.pollId,
        message.optionId,
      );
    } catch (error) {
      logger.error(error);
    }
  }
}
