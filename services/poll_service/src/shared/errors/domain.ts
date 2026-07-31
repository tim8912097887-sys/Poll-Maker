export abstract class DomainError extends Error {
  abstract readonly status: number;
  constructor(message: string) {
    super(message);
    this.name = new.target.name;
  }

  getStatus() {
    return this.status;
  }

  getResponse() {
    return this.message;
  }
}
