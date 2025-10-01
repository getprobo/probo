/**
 * @generated SignedSource<<b8a660825f6c3e3413e6505179ae3e1e>>
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
export type CreateCustomDomainDialogMutation$variables = {
  input: CreateCustomDomainInput;
};
export type CreateCustomDomainDialogMutation$data = {
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
      readonly verifiedAt: any | null | undefined;
    };
  };
};
export type CreateCustomDomainDialogMutation = {
  response: CreateCustomDomainDialogMutation$data;
  variables: CreateCustomDomainDialogMutation$variables;
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
            "name": "verifiedAt",
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
    "name": "CreateCustomDomainDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CreateCustomDomainDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9a6fad6d53811f74048b3ff0b64e0afe",
    "id": null,
    "metadata": {},
    "name": "CreateCustomDomainDialogMutation",
    "operationKind": "mutation",
    "text": "mutation CreateCustomDomainDialogMutation(\n  $input: CreateCustomDomainInput!\n) {\n  createCustomDomain(input: $input) {\n    customDomain {\n      id\n      domain\n      sslStatus\n      dnsRecords {\n        type\n        name\n        value\n        ttl\n        purpose\n      }\n      createdAt\n      updatedAt\n      verifiedAt\n      sslExpiresAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "312e66e894a78da19c28c7f03d03fe96";

export default node;
