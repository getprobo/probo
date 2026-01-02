/**
 * @generated SignedSource<<40e9cf14cc2a78785107ceb87a917437>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityGraphNodeQuery$variables = {
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityGraphNodeQuery$data = {
  readonly node: {
    readonly controlsInfo?: {
      readonly totalCount: number;
    };
    readonly createdAt?: any;
    readonly description?: string | null | undefined;
    readonly id?: string;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
    } | null | undefined;
    readonly snapshotId?: string | null | undefined;
    readonly sourceId?: string | null | undefined;
    readonly updatedAt?: any;
    readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityControlsTabFragment">;
  };
};
export type StateOfApplicabilityGraphNodeQuery = {
  response: StateOfApplicabilityGraphNodeQuery$data;
  variables: StateOfApplicabilityGraphNodeQuery$variables;
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
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "sourceId",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/)
  ],
  "storageKey": null
},
v10 = {
  "alias": "controlsInfo",
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 0
    }
  ],
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "totalCount",
      "storageKey": null
    }
  ],
  "storageKey": "controls(first:0)"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "StateOfApplicabilityGraphNodeQuery",
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
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "StateOfApplicabilityControlsTabFragment"
              }
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
    "name": "StateOfApplicabilityGraphNodeQuery",
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
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              {
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
                  (v3/*: any*/),
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
              }
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
    "cacheID": "d7931da7999de784b5056fefb8f49f59",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityGraphNodeQuery",
    "operationKind": "query",
    "text": "query StateOfApplicabilityGraphNodeQuery(\n  $stateOfApplicabilityId: ID!\n) {\n  node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      name\n      description\n      sourceId\n      snapshotId\n      createdAt\n      updatedAt\n      organization {\n        id\n      }\n      controlsInfo: controls(first: 0) {\n        totalCount\n      }\n      ...StateOfApplicabilityControlsTabFragment\n    }\n    id\n  }\n}\n\nfragment StateOfApplicabilityControlsTabFragment on StateOfApplicability {\n  id\n  availableControls {\n    controlId\n    sectionTitle\n    name\n    frameworkId\n    frameworkName\n    organizationId\n    stateOfApplicabilityId\n    state\n    exclusionJustification\n  }\n}\n"
  }
};
})();

(node as any).hash = "d5603b1a36e6c454e1436e655e4ec219";

export default node;
