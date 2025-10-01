/**
 * @generated SignedSource<<e1683e44614d1b0224fc8592dd72eb9f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
export type CreateCustomDomainInput = {
  domain: string;
  organizationId: string;
};
export type CustomDomainManagerCreateMutation$variables = {
  input: CreateCustomDomainInput;
};
export type CustomDomainManagerCreateMutation$data = {
  readonly createCustomDomain: {
    readonly customDomainEdge: {
      readonly node: {
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
        readonly isActive: boolean;
        readonly sslExpiresAt: any | null | undefined;
        readonly sslStatus: SSLStatus;
        readonly updatedAt: any;
        readonly verifiedAt: any | null | undefined;
      };
    };
  };
};
export type CustomDomainManagerCreateMutation = {
  response: CustomDomainManagerCreateMutation$data;
  variables: CustomDomainManagerCreateMutation$variables;
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
        "concreteType": "CustomDomainEdge",
        "kind": "LinkedField",
        "name": "customDomainEdge",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "CustomDomain",
            "kind": "LinkedField",
            "name": "node",
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
                "kind": "ScalarField",
                "name": "isActive",
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
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CustomDomainManagerCreateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CustomDomainManagerCreateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f10f04ab871fe2b42cc29c5789dc103a",
    "id": null,
    "metadata": {},
    "name": "CustomDomainManagerCreateMutation",
    "operationKind": "mutation",
    "text": "mutation CustomDomainManagerCreateMutation(\n  $input: CreateCustomDomainInput!\n) {\n  createCustomDomain(input: $input) {\n    customDomainEdge {\n      node {\n        id\n        domain\n        sslStatus\n        isActive\n        dnsRecords {\n          type\n          name\n          value\n          ttl\n          purpose\n        }\n        createdAt\n        updatedAt\n        verifiedAt\n        sslExpiresAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f2c706c80df1d0a9fa2e7ea4faa75fb3";

export default node;
