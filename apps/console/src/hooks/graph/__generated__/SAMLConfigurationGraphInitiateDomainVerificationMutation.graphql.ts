/**
 * @generated SignedSource<<55d58c6ee7e4dcd7ffe7e172aea932d5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type InitiateDomainVerificationInput = {
  emailDomain: string;
  organizationId: string;
};
export type SAMLConfigurationGraphInitiateDomainVerificationMutation$variables = {
  input: InitiateDomainVerificationInput;
};
export type SAMLConfigurationGraphInitiateDomainVerificationMutation$data = {
  readonly initiateDomainVerification: {
    readonly dnsRecord: string;
    readonly samlConfiguration: {
      readonly domainVerificationToken: string | null | undefined;
      readonly domainVerified: boolean;
      readonly domainVerifiedAt: any | null | undefined;
      readonly emailDomain: string;
      readonly id: string;
    };
  };
};
export type SAMLConfigurationGraphInitiateDomainVerificationMutation = {
  response: SAMLConfigurationGraphInitiateDomainVerificationMutation$data;
  variables: SAMLConfigurationGraphInitiateDomainVerificationMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "InitiateDomainVerificationPayload",
    "kind": "LinkedField",
    "name": "initiateDomainVerification",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SAMLConfiguration",
        "kind": "LinkedField",
        "name": "samlConfiguration",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "emailDomain",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "domainVerified",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "domainVerificationToken",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "domainVerifiedAt",
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "dnsRecord",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SAMLConfigurationGraphInitiateDomainVerificationMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphInitiateDomainVerificationMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "61453ca7fb2c7ff2fe8b657f6cfb7497",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphInitiateDomainVerificationMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphInitiateDomainVerificationMutation(\n  $input: InitiateDomainVerificationInput!\n) {\n  initiateDomainVerification(input: $input) {\n    samlConfiguration {\n      id\n      emailDomain\n      domainVerified\n      domainVerificationToken\n      domainVerifiedAt\n    }\n    dnsRecord\n  }\n}\n"
  }
};
})();

(node as any).hash = "5424f4fd94618327ae64292d31349e73";

export default node;
