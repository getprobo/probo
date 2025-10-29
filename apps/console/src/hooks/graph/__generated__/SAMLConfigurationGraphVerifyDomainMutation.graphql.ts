/**
 * @generated SignedSource<<8baf2363740e7eb47e7654f2f55c09df>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type VerifyDomainInput = {
  id: string;
};
export type SAMLConfigurationGraphVerifyDomainMutation$variables = {
  input: VerifyDomainInput;
};
export type SAMLConfigurationGraphVerifyDomainMutation$data = {
  readonly verifyDomain: {
    readonly samlConfiguration: {
      readonly domainVerified: boolean;
      readonly domainVerifiedAt: any | null | undefined;
      readonly id: string;
    };
    readonly verified: boolean;
  };
};
export type SAMLConfigurationGraphVerifyDomainMutation = {
  response: SAMLConfigurationGraphVerifyDomainMutation$data;
  variables: SAMLConfigurationGraphVerifyDomainMutation$variables;
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
    "concreteType": "VerifyDomainPayload",
    "kind": "LinkedField",
    "name": "verifyDomain",
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
            "name": "domainVerified",
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
        "name": "verified",
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
    "name": "SAMLConfigurationGraphVerifyDomainMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphVerifyDomainMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1f241529f518d2f0701d9c4f7a731d6b",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphVerifyDomainMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphVerifyDomainMutation(\n  $input: VerifyDomainInput!\n) {\n  verifyDomain(input: $input) {\n    samlConfiguration {\n      id\n      domainVerified\n      domainVerifiedAt\n    }\n    verified\n  }\n}\n"
  }
};
})();

(node as any).hash = "ea3f1fe691b0c36ff7e54663479f7e7c";

export default node;
