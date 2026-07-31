export const createPollKey = (pollId: string) => `poll:${pollId}:meta`;
export const createPollOptionsKey = (pollId: string) =>
  `poll:${pollId}:options`;
