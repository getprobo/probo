/**
 * @generated SignedSource<<1f66ba8e7d436aa543835a389d032651>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SettingsViewQuery$variables = {
  organizationID: string;
};
export type SettingsViewQuery$data = {
  readonly organization: {
    readonly aiFocused?: boolean | null | undefined;
    readonly companyType?: string | null | undefined;
    readonly connectors?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: string;
          readonly id: string;
          readonly name: string;
          readonly type: string;
        };
      }>;
    };
    readonly foundingYear?: number | null | undefined;
    readonly hasEnterpriseAccounts?: boolean | null | undefined;
    readonly hasRaisedMoney?: boolean | null | undefined;
    readonly id: string;
    readonly logoUrl?: string | null | undefined;
    readonly name?: string;
    readonly preMarketFit?: boolean | null | undefined;
    readonly users?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: string;
          readonly email: string;
          readonly fullName: string;
          readonly id: string;
        };
      }>;
    };
    readonly usesAiGeneratedCode?: boolean | null | undefined;
    readonly usesCloudProviders?: boolean | null | undefined;
    readonly vcBacked?: boolean | null | undefined;
  };
};
export type SettingsViewQuery = {
  response: SettingsViewQuery$data;
  variables: SettingsViewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationID"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationID"
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
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v6 = {
  "kind": "InlineFragment",
  "selections": [
    (v3/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "logoUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "foundingYear",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "companyType",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "preMarketFit",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "usesCloudProviders",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "aiFocused",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "usesAiGeneratedCode",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "vcBacked",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasRaisedMoney",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasEnterpriseAccounts",
      "storageKey": null
    },
    {
      "alias": null,
      "args": (v4/*: any*/),
      "concreteType": "UserConnection",
      "kind": "LinkedField",
      "name": "users",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "UserEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "User",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "fullName",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "email",
                  "storageKey": null
                },
                (v5/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "users(first:100)"
    },
    {
      "alias": null,
      "args": (v4/*: any*/),
      "concreteType": "ConnectorConnection",
      "kind": "LinkedField",
      "name": "connectors",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ConnectorEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Connector",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                (v3/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "type",
                  "storageKey": null
                },
                (v5/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "connectors(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SettingsViewQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v6/*: any*/)
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
    "name": "SettingsViewQuery",
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
          (v6/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "cc8d6ab0a62bf590970091613d1399d3",
    "id": null,
    "metadata": {},
    "name": "SettingsViewQuery",
    "operationKind": "query",
    "text": "query SettingsViewQuery(\n  $organizationID: ID!\n) {\n  organization: node(id: $organizationID) {\n    __typename\n    id\n    ... on Organization {\n      name\n      logoUrl\n      foundingYear\n      companyType\n      preMarketFit\n      usesCloudProviders\n      aiFocused\n      usesAiGeneratedCode\n      vcBacked\n      hasRaisedMoney\n      hasEnterpriseAccounts\n      users(first: 100) {\n        edges {\n          node {\n            id\n            fullName\n            email\n            createdAt\n          }\n        }\n      }\n      connectors(first: 100) {\n        edges {\n          node {\n            id\n            name\n            type\n            createdAt\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b9a0572ce518f5d473d09ef75ff2a80d";

export default node;
