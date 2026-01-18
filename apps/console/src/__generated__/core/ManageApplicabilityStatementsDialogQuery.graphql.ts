/**
 * @generated SignedSource<<d7bdd863ff08408724ca63ab5bf78393>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ManageApplicabilityStatementsDialogQuery$variables = {
  stateOfApplicabilityId: string;
};
export type ManageApplicabilityStatementsDialogQuery$data = {
  readonly node: {
    readonly availableControls?: ReadonlyArray<{
      readonly applicability: boolean | null | undefined;
      readonly applicabilityStatementId: string | null | undefined;
      readonly controlId: string;
      readonly frameworkId: string;
      readonly frameworkName: string;
      readonly justification: string | null | undefined;
      readonly name: string;
      readonly organizationId: string;
      readonly sectionTitle: string;
      readonly stateOfApplicabilityId: string | null | undefined;
    }>;
    readonly id?: string;
  };
};
export type ManageApplicabilityStatementsDialogQuery = {
  response: ManageApplicabilityStatementsDialogQuery$data;
  variables: ManageApplicabilityStatementsDialogQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "stateOfApplicabilityId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "stateOfApplicabilityId"
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
  "concreteType": "AvailableStateOfApplicabilityControl",
  "kind": "LinkedField",
  "name": "availableControls",
  "plural": true,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "controlId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "sectionTitle",
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
      "name": "frameworkId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "frameworkName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "organizationId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "applicabilityStatementId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "stateOfApplicabilityId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "applicability",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "justification",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ManageApplicabilityStatementsDialogQuery",
    "selections": [
      {
        "alias": null,
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
            "type": "StateOfApplicability",
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
    "name": "ManageApplicabilityStatementsDialogQuery",
    "selections": [
      {
        "alias": null,
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
            "type": "StateOfApplicability",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "3b03ae91880789d9a804d121f350ce7b",
    "id": null,
    "metadata": {},
    "name": "ManageApplicabilityStatementsDialogQuery",
    "operationKind": "query",
    "text": "query ManageApplicabilityStatementsDialogQuery(\n  $stateOfApplicabilityId: ID!\n) {\n  node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      availableControls {\n        controlId\n        sectionTitle\n        name\n        frameworkId\n        frameworkName\n        organizationId\n        applicabilityStatementId\n        stateOfApplicabilityId\n        applicability\n        justification\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "6eb3e0290f45063035657534b9040f12";

export default node;
