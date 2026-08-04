import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';

// Convert JS Date → Timestamp
export function toTimestamp(date: Date): Timestamp {
  const timestamp = new Timestamp();
  timestamp.setSeconds(Math.floor(date.getTime() / 1000));
  timestamp.setNanos((date.getTime() % 1000) * 1e6);
  return timestamp;
}
