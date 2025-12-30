/**
 * @generated SignedSource<<e47b433cc9fb58392ff933788bbe3deb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
export type CreateCustomDomainInput = {
  domain: string;
  organizationId: string;
};
export type NewDomainDialogMutation$variables = {
  input: CreateCustomDomainInput;
};
export type NewDomainDialogMutation$data = {
  readonly createCustomDomain: {
    readonly customDomain: {
      readonly createdAt: any;
      readonly dnsRecords: ReadonlyArray<{
        readonly name: string;
        readonly purpose: string;
        readonly ttl: number;
        readonly type: string;
        readonly value: string;
      }>;
      readonly domain: string;
      readonly id: string;
      readonly sslExpiresAt: any | null | undefined;
      readonly sslStatus: SSLStatus;
      readonly updatedAt: any;
    };
  };
};
export type NewDomainDialogMutation = {
  response: NewDomainDialogMutation$data;
  variables: NewDomainDialogMutation$variables;
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
    "concreteType": "CreateCustomDomainPayload",
    "kind": "LinkedField",
    "name": "createCustomDomain",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "CustomDomain",
        "kind": "LinkedField",
        "name": "customDomain",
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
            "name": "domain",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sslStatus",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "DNSRecordInstruction",
            "kind": "LinkedField",
            "name": "dnsRecords",
            "plural": true,
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "type",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "name",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "value",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "ttl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "purpose",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sslExpiresAt",
            "storageKey": null
          }
        ],
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
    "name": "NewDomainDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NewDomainDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b8aa27fe9ab3442dea9b96a70966524b",
    "id": null,
    "metadata": {},
    "name": "NewDomainDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewDomainDialogMutation(\n  $input: CreateCustomDomainInput!\n) {\n  createCustomDomain(input: $input) {\n    customDomain {\n      id\n      domain\n      sslStatus\n      dnsRecords {\n        type\n        name\n        value\n        ttl\n        purpose\n      }\n      createdAt\n      updatedAt\n      sslExpiresAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f49cd22c5d6656e662fad5afab6fc344";

export default node;
