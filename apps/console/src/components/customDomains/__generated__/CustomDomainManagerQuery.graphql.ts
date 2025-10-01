/**
 * @generated SignedSource<<04bd8d8635434cd05ea80edbb804b56f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
export type CustomDomainManagerQuery$variables = {
  organizationId: string;
};
export type CustomDomainManagerQuery$data = {
  readonly organization: {
    readonly customDomains?: {
      readonly edges: ReadonlyArray<{
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
      }>;
    };
    readonly id?: string;
  };
};
export type CustomDomainManagerQuery = {
  response: CustomDomainManagerQuery$data;
  variables: CustomDomainManagerQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 100
    }
  ],
  "concreteType": "CustomDomainConnection",
  "kind": "LinkedField",
  "name": "customDomains",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "CustomDomainEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "CustomDomain",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
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
  "storageKey": "customDomains(first:100)"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CustomDomainManagerQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/)
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CustomDomainManagerQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/)
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "c561676e49338e212721ed12e1069dd0",
    "id": null,
    "metadata": {},
    "name": "CustomDomainManagerQuery",
    "operationKind": "query",
    "text": "query CustomDomainManagerQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      customDomains(first: 100) {\n        edges {\n          node {\n            id\n            domain\n            sslStatus\n            isActive\n            dnsRecords {\n              type\n              name\n              value\n              ttl\n              purpose\n            }\n            createdAt\n            updatedAt\n            verifiedAt\n            sslExpiresAt\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "0aea6f600572fac571dbc0e1e6a3117c";

export default node;
