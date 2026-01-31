/**
 * @generated SignedSource<<2884e6469123701242f24cd275f8551b>>
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
export type NewCompliancePageDomainDialogMutation$variables = {
  input: CreateCustomDomainInput;
};
export type NewCompliancePageDomainDialogMutation$data = {
  readonly createCustomDomain: {
    readonly customDomain: {
      readonly canDelete: boolean;
      readonly createdAt: string;
      readonly dnsRecords: ReadonlyArray<{
        readonly name: string;
        readonly purpose: string;
        readonly ttl: number;
        readonly type: string;
        readonly value: string;
      }>;
      readonly domain: string;
      readonly id: string;
      readonly sslExpiresAt: string | null | undefined;
      readonly sslStatus: SSLStatus;
      readonly updatedAt: string;
    };
  };
};
export type NewCompliancePageDomainDialogMutation = {
  response: NewCompliancePageDomainDialogMutation$data;
  variables: NewCompliancePageDomainDialogMutation$variables;
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
          },
          {
            "alias": "canDelete",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:custom-domain:delete"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"core:custom-domain:delete\")"
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
    "name": "NewCompliancePageDomainDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NewCompliancePageDomainDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6f012080388b3c7dfd3bfc8c69b5d549",
    "id": null,
    "metadata": {},
    "name": "NewCompliancePageDomainDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewCompliancePageDomainDialogMutation(\n  $input: CreateCustomDomainInput!\n) {\n  createCustomDomain(input: $input) {\n    customDomain {\n      id\n      domain\n      sslStatus\n      dnsRecords {\n        type\n        name\n        value\n        ttl\n        purpose\n      }\n      createdAt\n      updatedAt\n      sslExpiresAt\n      canDelete: permission(action: \"core:custom-domain:delete\")\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7a1b76dbca18a3a0dbd2b1123e324b99";

export default node;
