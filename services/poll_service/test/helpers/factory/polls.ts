import { CreatePollType } from 'src/polls/schemas/create-poll.schema';

export const createPollSchema = (
  overrides?: Partial<CreatePollType>,
): CreatePollType => {
  const basePollSchema: CreatePollType = {
    isPrivate: false,
    title: 'Favorite Programming Language',
    options: [
      { optionText: 'TypeScript' },
      { optionText: 'Go' },
      { optionText: 'Rust' },
      { optionText: 'Python' },
    ],
    startedAt: new Date(
      new Date().getTime() + 24 * 60 * 60 * 1000,
    ).toISOString(),
    expiredAt: new Date(
      new Date().getTime() + 48 * 60 * 60 * 1000,
    ).toISOString(),
  };
  return {
    ...basePollSchema,
    ...overrides,
  };
};
