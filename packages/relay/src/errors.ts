export class UnAuthenticatedError extends Error {
  constructor(message?: string) {
    super(message || "UNAUTHENTICATED");
    this.name = "UnAuthenticatedError";
    Object.setPrototypeOf(this, UnAuthenticatedError.prototype);
  }
}

export class InternalServerError extends Error {
  constructor() {
    super("INTERNAL_SERVER_ERROR");
    this.name = "InternalServerError";
    Object.setPrototypeOf(this, InternalServerError.prototype);
  }
}

export class AuthenticationRequiredError extends Error {
  public redirectUrl: string;
  public requiresSaml: boolean;
  public organizationId: string;
  public samlConfigId?: string;

  constructor(extensions: {
    redirectUrl: string;
    requiresSaml: boolean;
    organizationId: string;
    samlConfigId?: string;
  }) {
    super("AUTHENTICATION_REQUIRED");
    this.name = "AuthenticationRequiredError";
    Object.setPrototypeOf(this, AuthenticationRequiredError.prototype);
    this.redirectUrl = extensions.redirectUrl;
    this.requiresSaml = extensions.requiresSaml;
    this.organizationId = extensions.organizationId;
    this.samlConfigId = extensions.samlConfigId;
  }
}

export class UnauthorizedError extends Error {
  constructor(message?: string) {
    super(message || "UNAUTHORIZED");
    this.name = "UnauthorizedError";
    Object.setPrototypeOf(this, UnauthorizedError.prototype);
  }
}

export class ForbiddenError extends Error {
  constructor(message?: string) {
    super(message || "FORBIDDEN");
    this.name = "ForbiddenError";
    Object.setPrototypeOf(this, ForbiddenError.prototype);
  }
}
