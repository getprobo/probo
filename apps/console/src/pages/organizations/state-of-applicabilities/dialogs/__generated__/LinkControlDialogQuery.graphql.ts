/**
 * @generated SignedSource<<930083c6b44cb4fac44938e6422415c6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type StateOfApplicabilityControlState = "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED";
export type LinkControlDialogQuery$variables = {
  stateOfApplicabilityId: string;
};
export type LinkControlDialogQuery$data = {
  readonly node: {
    readonly availableControls?: ReadonlyArray<{
      readonly controlId: string;
      readonly exclusionJustification: string | null | undefined;
      readonly frameworkId: string;
      readonly frameworkName: string;
      readonly name: string;
      readonly organizationId: string;
      readonly sectionTitle: string;
      readonly state: StateOfApplicabilityControlState | null | undefined;
      readonly stateOfApplicabilityId: string | null | undefined;
    }>;
    readonly id?: string;
  };
};
export type LinkControlDialogQuery = {
  response: LinkControlDialogQuery$data;
  variables: LinkControlDialogQuery$variables;
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
  "concreteType": "AvailableControlForStateOfApplicability",
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
      "name": "stateOfApplicabilityId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "exclusionJustification",
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
    "name": "LinkControlDialogQuery",
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
    "name": "LinkControlDialogQuery",
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
    "cacheID": "db215e59c8633cdeecbe7ec9a6431352",
    "id": null,
    "metadata": {},
    "name": "LinkControlDialogQuery",
    "operationKind": "query",
    "text": "query LinkControlDialogQuery(\n  $stateOfApplicabilityId: ID!\n) {\n  node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      availableControls {\n        controlId\n        sectionTitle\n        name\n        frameworkId\n        frameworkName\n        organizationId\n        stateOfApplicabilityId\n        state\n        exclusionJustification\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "dcfac7b8bc44fe2e92c36751d5f4734c";

export default node;
