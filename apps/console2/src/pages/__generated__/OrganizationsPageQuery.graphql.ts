/**
 * @generated SignedSource<<91cc89e5e07572ab16c7bc9b8443fcd6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationsPageQuery$variables = Record<PropertyKey, never>;
export type OrganizationsPageQuery$data = {
  readonly viewer: {
    readonly organizations: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly logoUrl: string | null | undefined;
          readonly name: string;
        };
      }>;
    };
  };
};
export type OrganizationsPageQuery = {
  response: OrganizationsPageQuery$data;
  variables: OrganizationsPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 25
    }
  ],
  "concreteType": "OrganizationConnection",
  "kind": "LinkedField",
  "name": "organizations",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "OrganizationEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "Organization",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v0/*: any*/),
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
              "name": "logoUrl",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "organizations(first:25)"
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationsPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationsPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          (v0/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "e771f8e44e9f9224d9cbe0f271b096ed",
    "id": null,
    "metadata": {},
    "name": "OrganizationsPageQuery",
    "operationKind": "query",
    "text": "query OrganizationsPageQuery {\n  viewer {\n    organizations(first: 25) {\n      edges {\n        node {\n          id\n          name\n          logoUrl\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "457b1b8b28355830d77458bab7af23f2";

export default node;
