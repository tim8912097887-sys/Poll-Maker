export type VoteCreatedMessage = {
  eventId: string;
  pollId: string;
  optionId: string;
  votedAt: string;
};
