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
  async getPolls(): Promise<GetPollsDto> {
    const polls = await this.pollsService.getPublicPolls();
    return GetPollsDto.toDto(polls);
  }

  @Post()
  async createPoll(
    @Body(new ZodValidationPipe(CreatePollSchema))
    createPollSchema: CreatePollType,
  ): Promise<CreatePollDto> {
    const poll = await this.pollsService.createPoll(createPollSchema);
    return CreatePollDto.toDto(poll);
  }

  @Delete(':id')
  async deletePoll(
    @Param('id', new ParseUUIDPipe()) id: string,
  ): Promise<string> {
    await this.pollsService.deletePoll(id);
    return 'Successfully deleted poll';
  }
}
