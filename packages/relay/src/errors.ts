export class UnAuthenticatedError extends Error {
  constructor(message?: string) {
    super(message || "UNAUTHENTICATED");
    this.name = "UnAuthenticatedError";
    Object.setPrototypeOf(this, UnAuthenticatedError.prototype);
  }
}

export class FullNameRequiredError extends Error {
  constructor(message?: string) {
    super(message || "FULL_NAME_REQUIRED");
    this.name = "FullNameRequiredError";
    Object.setPrototypeOf(this, FullNameRequiredError.prototype);
  }
}

export class NDASignatureRequiredError extends Error {
  constructor(message?: string) {
    super(message || "NDA_SIGNATURE_REQUIRED");
    this.name = "NDASignatureRequiredError";
    Object.setPrototypeOf(this, NDASignatureRequiredError.prototype);
  }
}

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
    Object.setPrototypeOf(this, InternalServerError.prototype);
  }
}

export class AssumptionRequiredError extends Error {
  constructor(message?: string) {
    super(message ?? "ASSUMPTION_REQUIRED");
    this.name = "AssumptionRequiredError";
    Object.setPrototypeOf(this, AssumptionRequiredError.prototype)
  }
}

export class ForbiddenError extends Error {
  constructor(message?: string) {
    super(message || "FORBIDDEN");
    this.name = "ForbiddenError";
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}
