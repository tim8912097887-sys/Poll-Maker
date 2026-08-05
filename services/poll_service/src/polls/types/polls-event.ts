export type VoteCreatedMessage = {
  eventId: string;
  pollId: string;
  optionId: string;
  votedAt: string;
};

export type VoteCreatedEvent = {
  key: string;
  value: VoteCreatedMessage;
};
