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

@Controller('/api/v1/polls')
export class PollsController {
  constructor(private readonly pollsService: PollsService) {}

  @Get()
  getPolls(): GetPollsDto {
    const polls = this.pollsService.getPolls();
    return GetPollsDto.toDto(polls);
  }

  @Post()
  createPoll(
    @Body(new ZodValidationPipe(CreatePollSchema))
    createPollSchema: CreatePollType,
  ): CreatePollDto {
    const poll = this.pollsService.createPoll(createPollSchema);
    return CreatePollDto.toDto(poll);
  }

  @Delete(':id')
  deletePoll(@Param('id', new ParseUUIDPipe()) id: string): string {
    this.pollsService.deletePoll(id);
    return 'Successfully deleted poll';
  }
}
