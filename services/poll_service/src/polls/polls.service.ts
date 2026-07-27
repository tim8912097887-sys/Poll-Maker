import { Injectable, NotFoundException } from '@nestjs/common';
import { PollType } from './schemas/polls.schema';
import { PollOptionType } from './schemas/poll-options.schema';
import { CreatePollResponse, GetPollsResponse } from './types/polls-response';
import { CreatePollType } from './schemas/create-poll.schema';

export const samplePolls: PollType[] = [
  {
    id: 'c1a1f3d2-9b4e-4a7a-8b1a-1f2d3e4f5a6b',
    title: 'Best Programming Language',
    isPrivate: false,
    creatorSession: 'session_001',
    createdAt: new Date('2026-07-01T10:00:00Z'),
    startedAt: new Date('2026-07-02T08:00:00Z'),
    expiredAt: new Date('2026-07-10T08:00:00Z'),
  },
  {
    id: 'd2b2f4e3-8c5f-4b8a-9c2b-2f3e4f5a6b7c',
    title: 'Favorite Movie Genre',
    isPrivate: false,
    creatorSession: 'session_002',
    createdAt: new Date('2026-07-03T09:30:00Z'),
    startedAt: new Date('2026-07-04T09:00:00Z'),
    expiredAt: new Date('2026-07-12T09:00:00Z'),
  },
  {
    id: 'e3c3f5a4-7d6e-4c9b-8d3c-3f4e5a6b7c8d',
    title: 'Best Soccer Player',
    isPrivate: true,
    creatorSession: 'session_003',
    createdAt: new Date('2026-07-05T11:00:00Z'),
    startedAt: new Date('2026-07-06T11:00:00Z'),
    expiredAt: new Date('2026-07-15T11:00:00Z'),
  },
  {
    id: 'f4d4a6b5-6e7f-4d0c-9e4d-4f5a6b7c8d9e',
    title: 'Preferred Frontend Framework',
    isPrivate: false,
    creatorSession: 'session_004',
    createdAt: new Date('2026-07-06T12:00:00Z'),
    startedAt: new Date('2026-07-07T12:00:00Z'),
    expiredAt: new Date('2026-07-14T12:00:00Z'),
  },
  {
    id: 'a5e5b7c6-5f8g-4e1d-9f5e-5a6b7c8d9e0f',
    title: 'Best NBA Team',
    isPrivate: true,
    creatorSession: 'session_005',
    createdAt: new Date('2026-07-07T13:00:00Z'),
    startedAt: new Date('2026-07-08T13:00:00Z'),
    expiredAt: new Date('2026-07-16T13:00:00Z'),
  },
  {
    id: 'b6f6c8d7-4g9h-4f2e-9g6f-6b7c8d9e0f1a',
    title: 'Favorite Music Artist',
    isPrivate: false,
    creatorSession: 'session_006',
    createdAt: new Date('2026-07-08T14:00:00Z'),
    startedAt: new Date('2026-07-09T14:00:00Z'),
    expiredAt: new Date('2026-07-17T14:00:00Z'),
  },
  {
    id: 'c7g7d9e8-3h0i-4g3f-9h7g-7c8d9e0f1a2b',
    title: 'Best Cloud Provider',
    isPrivate: false,
    creatorSession: 'session_007',
    createdAt: new Date('2026-07-09T15:00:00Z'),
    startedAt: new Date('2026-07-10T15:00:00Z'),
    expiredAt: new Date('2026-07-18T15:00:00Z'),
  },
  {
    id: 'd8h8e0f9-2i1j-4h4g-9i8h-8d9e0f1a2b3c',
    title: 'Preferred Database',
    isPrivate: true,
    creatorSession: 'session_008',
    createdAt: new Date('2026-07-10T16:00:00Z'),
    startedAt: new Date('2026-07-11T16:00:00Z'),
    expiredAt: new Date('2026-07-19T16:00:00Z'),
  },
  {
    id: 'e9i9f1a0-1j2k-4i5h-9j9i-9e0f1a2b3c4d',
    title: 'Best Streaming Platform',
    isPrivate: false,
    creatorSession: 'session_009',
    createdAt: new Date('2026-07-11T17:00:00Z'),
    startedAt: new Date('2026-07-12T17:00:00Z'),
    expiredAt: new Date('2026-07-20T17:00:00Z'),
  },
  {
    id: 'f0j0a2b1-0k3l-4j6i-9k0j-0f1a2b3c4d5e',
    title: 'Best AI Model',
    isPrivate: false,
    creatorSession: 'session_010',
    createdAt: new Date('2026-07-12T18:00:00Z'),
    startedAt: new Date('2026-07-13T18:00:00Z'),
    expiredAt: new Date('2026-07-21T18:00:00Z'),
  },
  {
    id: 'a1k1b3c2-9l4m-4k7j-9l1k-1a2b3c4d5e6f',
    title: 'Favorite Vacation Destination',
    isPrivate: true,
    creatorSession: 'session_011',
    createdAt: new Date('2026-07-13T19:00:00Z'),
    startedAt: new Date('2026-07-14T19:00:00Z'),
    expiredAt: new Date('2026-07-22T19:00:00Z'),
  },
  {
    id: 'b2l2c4d3-8m5n-4l8k-9m2l-2b3c4d5e6f7g',
    title: 'Best Mobile OS',
    isPrivate: false,
    creatorSession: 'session_012',
    createdAt: new Date('2026-07-14T20:00:00Z'),
    startedAt: new Date('2026-07-15T20:00:00Z'),
    expiredAt: new Date('2026-07-23T20:00:00Z'),
  },
  //   {
  //     id: 'c3m3d5e4-7n6o-4m9l-9n3m-3c4d5e6f7g8h',
  //     title: 'Best Coffee Brand',
  //     isPrivate: false,
  //     creatorSession: 'session_013',
  //     createdAt: new Date('2026-07-15T21:00:00Z'),
  //     startedAt: new Date('2026-07-16T21:00:00Z'),
  //     expiredAt: new Date('2026-07-24T21:00:00Z'),
  //   },
  //   {
  //     id: 'd4n4e6f5-6o7p-4n0m-9o4n-4d5e6f7g8h9i',
  //     title: 'Best Smartphone Brand',
  //     isPrivate: true,
  //     creatorSession: 'session_014',
  //     createdAt: new Date('2026-07-16T22:00:00Z'),
  //     startedAt: new Date('2026-07-17T22:00:00Z'),
  //     expiredAt: new Date('2026-07-25T22:00:00Z'),
  //   },
  //   {
  //     id: 'e5o5f7g6-5p8q-4o1n-9p5o-5e6f7g8h9i0j',
  //     title: 'Best Coding Editor',
  //     isPrivate: false,
  //     creatorSession: 'session_015',
  //     createdAt: new Date('2026-07-17T23:00:00Z'),
  //     startedAt: new Date('2026-07-18T23:00:00Z'),
  //     expiredAt: new Date('2026-07-26T23:00:00Z'),
  //   },
];

export const samplePollOptions: PollOptionType[] = [
  // Poll 1: Best Programming Language
  {
    id: '11111111-aaaa-bbbb-cccc-000000000001',
    pollId: 'c1a1f3d2-9b4e-4a7a-8b1a-1f2d3e4f5a6b',
    optionText: 'JavaScript',
    voteCounts: 0,
    createdAt: new Date('2026-07-02T08:00:00Z'),
  },
  {
    id: '11111111-aaaa-bbbb-cccc-000000000002',
    pollId: 'c1a1f3d2-9b4e-4a7a-8b1a-1f2d3e4f5a6b',
    optionText: 'Python',
    voteCounts: 0,
    createdAt: new Date('2026-07-02T08:00:00Z'),
  },
  {
    id: '11111111-aaaa-bbbb-cccc-000000000003',
    pollId: 'c1a1f3d2-9b4e-4a7a-8b1a-1f2d3e4f5a6b',
    optionText: 'Go',
    voteCounts: 0,
    createdAt: new Date('2026-07-02T08:00:00Z'),
  },

  // Poll 2: Favorite Movie Genre
  {
    id: '22222222-aaaa-bbbb-cccc-000000000001',
    pollId: 'd2b2f4e3-8c5f-4b8a-9c2b-2f3e4f5a6b7c',
    optionText: 'Action',
    voteCounts: 0,
    createdAt: new Date('2026-07-04T09:00:00Z'),
  },
  {
    id: '22222222-aaaa-bbbb-cccc-000000000002',
    pollId: 'd2b2f4e3-8c5f-4b8a-9c2b-2f3e4f5a6b7c',
    optionText: 'Comedy',
    voteCounts: 0,
    createdAt: new Date('2026-07-04T09:00:00Z'),
  },
  {
    id: '22222222-aaaa-bbbb-cccc-000000000003',
    pollId: 'd2b2f4e3-8c5f-4b8a-9c2b-2f3e4f5a6b7c',
    optionText: 'Drama',
    voteCounts: 0,
    createdAt: new Date('2026-07-04T09:00:00Z'),
  },

  // Poll 3: Best Soccer Player
  {
    id: '33333333-aaaa-bbbb-cccc-000000000001',
    pollId: 'e3c3f5a4-7d6e-4c9b-8d3c-3f4e5a6b7c8d',
    optionText: 'Lionel Messi',
    voteCounts: 0,
    createdAt: new Date('2026-07-06T11:00:00Z'),
  },
  {
    id: '33333333-aaaa-bbbb-cccc-000000000002',
    pollId: 'e3c3f5a4-7d6e-4c9b-8d3c-3f4e5a6b7c8d',
    optionText: 'Cristiano Ronaldo',
    voteCounts: 0,
    createdAt: new Date('2026-07-06T11:00:00Z'),
  },
  {
    id: '33333333-aaaa-bbbb-cccc-000000000003',
    pollId: 'e3c3f5a4-7d6e-4c9b-8d3c-3f4e5a6b7c8d',
    optionText: 'Erling Haaland',
    voteCounts: 0,
    createdAt: new Date('2026-07-06T11:00:00Z'),
  },

  // Poll 4: Preferred Frontend Framework
  {
    id: '44444444-aaaa-bbbb-cccc-000000000001',
    pollId: 'f4d4a6b5-6e7f-4d0c-9e4d-4f5a6b7c8d9e',
    optionText: 'React',
    voteCounts: 0,
    createdAt: new Date('2026-07-07T12:00:00Z'),
  },
  {
    id: '44444444-aaaa-bbbb-cccc-000000000002',
    pollId: 'f4d4a6b5-6e7f-4d0c-9e4d-4f5a6b7c8d9e',
    optionText: 'Vue',
    voteCounts: 0,
    createdAt: new Date('2026-07-07T12:00:00Z'),
  },
  {
    id: '44444444-aaaa-bbbb-cccc-000000000003',
    pollId: 'f4d4a6b5-6e7f-4d0c-9e4d-4f5a6b7c8d9e',
    optionText: 'Angular',
    voteCounts: 0,
    createdAt: new Date('2026-07-07T12:00:00Z'),
  },

  // Poll 5: Best NBA Team
  {
    id: '55555555-aaaa-bbbb-cccc-000000000001',
    pollId: 'a5e5b7c6-5f8g-4e1d-9f5e-5a6b7c8d9e0f',
    optionText: 'Los Angeles Lakers',
    voteCounts: 0,
    createdAt: new Date('2026-07-08T13:00:00Z'),
  },
  {
    id: '55555555-aaaa-bbbb-cccc-000000000002',
    pollId: 'a5e5b7c6-5f8g-4e1d-9f5e-5a6b7c8d9e0f',
    optionText: 'Chicago Bulls',
    voteCounts: 0,
    createdAt: new Date('2026-07-08T13:00:00Z'),
  },
  {
    id: '55555555-aaaa-bbbb-cccc-000000000003',
    pollId: 'a5e5b7c6-5f8g-4e1d-9f5e-5a6b7c8d9e0f',
    optionText: 'Golden State Warriors',
    voteCounts: 0,
    createdAt: new Date('2026-07-08T13:00:00Z'),
  },
  // Poll 6: Favorite Music Artist
  {
    id: '66666666-aaaa-bbbb-cccc-000000000001',
    pollId: 'b6f6c8d7-4g9h-4f2e-9g6f-6b7c8d9e0f1a',
    optionText: 'BTS',
    voteCounts: 0,
    createdAt: new Date('2026-07-09T14:00:00Z'),
  },
  {
    id: '66666666-aaaa-bbbb-cccc-000000000002',
    pollId: 'b6f6c8d7-4g9h-4f2e-9g6f-6b7c8d9e0f1a',
    optionText: 'Madonna',
    voteCounts: 0,
    createdAt: new Date('2026-07-09T14:00:00Z'),
  },
  {
    id: '66666666-aaaa-bbbb-cccc-000000000003',
    pollId: 'b6f6c8d7-4g9h-4f2e-9g6f-6b7c8d9e0f1a',
    optionText: 'Rainie Yang',
    voteCounts: 0,
    createdAt: new Date('2026-07-09T14:00:00Z'),
  },

  // Poll 7: Best Cloud Provider
  {
    id: '77777777-aaaa-bbbb-cccc-000000000001',
    pollId: 'c7g7d9e8-3h0i-4g3f-9h7g-7c8d9e0f1a2b',
    optionText: 'AWS',
    voteCounts: 0,
    createdAt: new Date('2026-07-10T15:00:00Z'),
  },
  {
    id: '77777777-aaaa-bbbb-cccc-000000000002',
    pollId: 'c7g7d9e8-3h0i-4g3f-9h7g-7c8d9e0f1a2b',
    optionText: 'Azure',
    voteCounts: 0,
    createdAt: new Date('2026-07-10T15:00:00Z'),
  },
  {
    id: '77777777-aaaa-bbbb-cccc-000000000003',
    pollId: 'c7g7d9e8-3h0i-4g3f-9h7g-7c8d9e0f1a2b',
    optionText: 'Google Cloud',
    voteCounts: 0,
    createdAt: new Date('2026-07-10T15:00:00Z'),
  },

  // Poll 8: Preferred Database
  {
    id: '88888888-aaaa-bbbb-cccc-000000000001',
    pollId: 'd8h8e0f9-2i1j-4h4g-9i8h-8d9e0f1a2b3c',
    optionText: 'PostgreSQL',
    voteCounts: 0,
    createdAt: new Date('2026-07-11T16:00:00Z'),
  },
  {
    id: '88888888-aaaa-bbbb-cccc-000000000002',
    pollId: 'd8h8e0f9-2i1j-4h4g-9i8h-8d9e0f1a2b3c',
    optionText: 'MySQL',
    voteCounts: 0,
    createdAt: new Date('2026-07-11T16:00:00Z'),
  },
  {
    id: '88888888-aaaa-bbbb-cccc-000000000003',
    pollId: 'd8h8e0f9-2i1j-4h4g-9i8h-8d9e0f1a2b3c',
    optionText: 'MongoDB',
    voteCounts: 0,
    createdAt: new Date('2026-07-11T16:00:00Z'),
  },

  // Poll 9: Best Streaming Platform
  {
    id: '99999999-aaaa-bbbb-cccc-000000000001',
    pollId: 'e9i9f1a0-1j2k-4i5h-9j9i-9e0f1a2b3c4d',
    optionText: 'Netflix',
    voteCounts: 0,
    createdAt: new Date('2026-07-12T17:00:00Z'),
  },
  {
    id: '99999999-aaaa-bbbb-cccc-000000000002',
    pollId: 'e9i9f1a0-1j2k-4i5h-9j9i-9e0f1a2b3c4d',
    optionText: 'Disney+',
    voteCounts: 0,
    createdAt: new Date('2026-07-12T17:00:00Z'),
  },
  {
    id: '99999999-aaaa-bbbb-cccc-000000000003',
    pollId: 'e9i9f1a0-1j2k-4i5h-9j9i-9e0f1a2b3c4d',
    optionText: 'Amazon Prime Video',
    voteCounts: 0,
    createdAt: new Date('2026-07-12T17:00:00Z'),
  },

  // Poll 10: Best AI Model
  {
    id: '10101010-aaaa-bbbb-cccc-000000000001',
    pollId: 'f0j0a2b1-0k3l-4j6i-9k0j-0f1a2b3c4d5e',
    optionText: 'GPT-4',
    voteCounts: 0,
    createdAt: new Date('2026-07-13T18:00:00Z'),
  },
  {
    id: '10101010-aaaa-bbbb-cccc-000000000002',
    pollId: 'f0j0a2b1-0k3l-4j6i-9k0j-0f1a2b3c4d5e',
    optionText: 'Claude',
    voteCounts: 0,
    createdAt: new Date('2026-07-13T18:00:00Z'),
  },
  {
    id: '10101010-aaaa-bbbb-cccc-000000000003',
    pollId: 'f0j0a2b1-0k3l-4j6i-9k0j-0f1a2b3c4d5e',
    optionText: 'Gemini',
    voteCounts: 0,
    createdAt: new Date('2026-07-13T18:00:00Z'),
  },

  // Poll 11: Favorite Vacation Destination
  {
    id: '11111112-aaaa-bbbb-cccc-000000000001',
    pollId: 'a1k1b3c2-9l4m-4k7j-9l1k-1a2b3c4d5e6f',
    optionText: 'Paris',
    voteCounts: 0,
    createdAt: new Date('2026-07-14T19:00:00Z'),
  },
  {
    id: '11111112-aaaa-bbbb-cccc-000000000002',
    pollId: 'a1k1b3c2-9l4m-4k7j-9l1k-1a2b3c4d5e6f',
    optionText: 'Tokyo',
    voteCounts: 0,
    createdAt: new Date('2026-07-14T19:00:00Z'),
  },
  {
    id: '11111112-aaaa-bbbb-cccc-000000000003',
    pollId: 'a1k1b3c2-9l4m-4k7j-9l1k-1a2b3c4d5e6f',
    optionText: 'New York',
    voteCounts: 0,
    createdAt: new Date('2026-07-14T19:00:00Z'),
  },

  // Poll 12: Best Mobile OS
  {
    id: '12121212-aaaa-bbbb-cccc-000000000001',
    pollId: 'b2l2c4d3-8m5n-4l8k-9m2l-2b3c4d5e6f7g',
    optionText: 'iOS',
    voteCounts: 0,
    createdAt: new Date('2026-07-15T20:00:00Z'),
  },
  {
    id: '12121212-aaaa-bbbb-cccc-000000000002',
    pollId: 'b2l2c4d3-8m5n-4l8k-9m2l-2b3c4d5e6f7g',
    optionText: 'Android',
    voteCounts: 0,
    createdAt: new Date('2026-07-15T20:00:00Z'),
  },
];

@Injectable()
export class PollsService {
  getPolls(): GetPollsResponse {
    const returnPolls = samplePolls.map((poll) => ({
      id: poll.id,
      title: poll.title,
      isPrivate: poll.isPrivate,
      creatorSession: poll.creatorSession,
      startedAt: poll.startedAt,
      expiredAt: poll.expiredAt,
      options: samplePollOptions
        .filter((option) => option.pollId === poll.id)
        .map((option) => ({
          optionText: option.optionText,
          voteCounts: option.voteCounts,
        })),
    }));

    return returnPolls.filter((poll) => poll.isPrivate === false);
  }

  createPoll(createPollSchema: CreatePollType): CreatePollResponse {
    const pollId = crypto.randomUUID();
    const creatorSession = crypto.randomUUID();
    const { options, startedAt, expiredAt, ...rest } = createPollSchema;
    const poll = {
      ...rest,
      id: pollId,
      creatorSession: creatorSession,
      createdAt: new Date(),
      startedAt: new Date(startedAt),
      expiredAt: new Date(expiredAt),
    };
    const pollOptions = options.map((option) => ({
      ...option,
      id: crypto.randomUUID(),
      pollId: pollId,
      voteCounts: 0,
      createdAt: new Date(),
    }));

    samplePolls.push(poll);
    samplePollOptions.push(...pollOptions);

    const response: CreatePollResponse = {
      ...poll,
      options: pollOptions,
    };
    return response;
  }

  deletePoll(id: string): string {
    const pollNumber = samplePolls.findIndex((poll) => poll.id === id);
    if (pollNumber === -1) throw new NotFoundException('Poll not found');
    samplePolls.splice(pollNumber, 1);
    return `This action removes a #${id} poll`;
  }
}
