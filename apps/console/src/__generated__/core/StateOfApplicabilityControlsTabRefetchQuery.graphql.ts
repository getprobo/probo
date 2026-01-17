/**
 * @generated SignedSource<<82aac43cc256062befe9f2126eb20c4e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityControlsTabRefetchQuery$variables = {
  id: string;
};
export type StateOfApplicabilityControlsTabRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityControlsTabFragment">;
  };
};
export type StateOfApplicabilityControlsTabRefetchQuery = {
  response: StateOfApplicabilityControlsTabRefetchQuery$data;
  variables: StateOfApplicabilityControlsTabRefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "id"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "StateOfApplicabilityControlsTabRefetchQuery",
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
            "args": null,
            "kind": "FragmentSpread",
            "name": "StateOfApplicabilityControlsTabFragment"
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
    "name": "StateOfApplicabilityControlsTabRefetchQuery",
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
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
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
                    "value": "core:state-of-applicability-control-mapping:create"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:create\")"
              },
              {
                "alias": "canDeleteStateOfApplicabilityControlMapping",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:state-of-applicability-control-mapping:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:delete\")"
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
    "cacheID": "fd5b512b7274ba6a3047960c796c82bf",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityControlsTabRefetchQuery",
    "operationKind": "query",
    "text": "query StateOfApplicabilityControlsTabRefetchQuery(\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...StateOfApplicabilityControlsTabFragment\n    id\n  }\n}\n\nfragment StateOfApplicabilityControlsTabFragment on StateOfApplicability {\n  id\n  controlsInfo: controls(first: 0) {\n    totalCount\n  }\n  canCreateStateOfApplicabilityControlMapping: permission(action: \"core:state-of-applicability-control-mapping:create\")\n  canDeleteStateOfApplicabilityControlMapping: permission(action: \"core:state-of-applicability-control-mapping:delete\")\n  availableControls {\n    controlId\n    sectionTitle\n    name\n    frameworkId\n    frameworkName\n    organizationId\n    stateOfApplicabilityId\n    applicability\n    justification\n    bestPractice\n    regulatory\n    contractual\n    riskAssessment\n  }\n}\n"
  }
};
})();

(node as any).hash = "37dd1660caf4dea39bb697b99a8bf511";

export default node;
