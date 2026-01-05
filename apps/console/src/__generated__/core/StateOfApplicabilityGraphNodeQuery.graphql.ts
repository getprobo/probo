/**
 * @generated SignedSource<<74ef0d7b05d90eaaac07851cfb57409f>>
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
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly createdAt?: string;
    readonly id?: string;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
    } | null | undefined;
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly snapshotId?: string | null | undefined;
    readonly sourceId?: string | null | undefined;
    readonly updatedAt?: string;
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
  "name": "sourceId",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v8 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:update\")"
},
v9 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:delete\")"
},
v10 = {
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
v11 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fullName",
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
              (v11/*: any*/),
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
              (v11/*: any*/),
              {
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
              },
              {
                "alias": "canCreateStateOfApplicabilityControlMapping",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:state-of-applicability:control-mapping:create"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability:control-mapping:create\")"
              },
              {
                "alias": "canDeleteStateOfApplicabilityControlMapping",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:state-of-applicability:control-mapping:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability:control-mapping:delete\")"
              },
              {
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
                    "name": "applicability",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "justification",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "bestPractice",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "regulatory",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "contractual",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "riskAssessment",
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
    "cacheID": "dfdfadde088e82025aaa6c9c290f96b6",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityGraphNodeQuery",
    "operationKind": "query",
    "text": "query StateOfApplicabilityGraphNodeQuery(\n  $stateOfApplicabilityId: ID!\n) {\n  node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      name\n      sourceId\n      snapshotId\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:state-of-applicability:update\")\n      canDelete: permission(action: \"core:state-of-applicability:delete\")\n      organization {\n        id\n      }\n      owner {\n        id\n        fullName\n      }\n      ...StateOfApplicabilityControlsTabFragment\n    }\n    id\n  }\n}\n\nfragment StateOfApplicabilityControlsTabFragment on StateOfApplicability {\n  id\n  controlsInfo: controls(first: 0) {\n    totalCount\n  }\n  canCreateStateOfApplicabilityControlMapping: permission(action: \"core:state-of-applicability:control-mapping:create\")\n  canDeleteStateOfApplicabilityControlMapping: permission(action: \"core:state-of-applicability:control-mapping:delete\")\n  availableControls {\n    controlId\n    sectionTitle\n    name\n    frameworkId\n    frameworkName\n    organizationId\n    stateOfApplicabilityId\n    applicability\n    justification\n    bestPractice\n    regulatory\n    contractual\n    riskAssessment\n  }\n}\n"
  }
};
})();

(node as any).hash = "1a1f3480b9c66597d146c48ed1b00bd1";

export default node;
