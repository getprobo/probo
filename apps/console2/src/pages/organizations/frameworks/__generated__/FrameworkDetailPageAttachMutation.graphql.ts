/**
 * @generated SignedSource<<81d008fc4116a78a92e2d925f8d6daa7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlMeasureMappingInput = {
  controlId: string;
  measureId: string;
};
export type FrameworkDetailPageAttachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlMeasureMappingInput;
};
export type FrameworkDetailPageAttachMutation$data = {
  readonly createControlMeasureMapping: {
    readonly measureEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedMeasuresCardFragment">;
      };
    };
  };
};
export type FrameworkDetailPageAttachMutation = {
  response: FrameworkDetailPageAttachMutation$data;
  variables: FrameworkDetailPageAttachMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkDetailPageAttachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createControlMeasureMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "MeasureEdge",
            "kind": "LinkedField",
            "name": "measureEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Measure",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedMeasuresCardFragment"
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
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "FrameworkDetailPageAttachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "createControlMeasureMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "MeasureEdge",
            "kind": "LinkedField",
            "name": "measureEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Measure",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
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
                    "name": "state",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "measureEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "9fcb599cd78f3108d6d71aa571d7005d",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageAttachMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageAttachMutation(\n  $input: CreateControlMeasureMappingInput!\n) {\n  createControlMeasureMapping(input: $input) {\n    measureEdge {\n      node {\n        id\n        ...LinkedMeasuresCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n"
  }
};
})();

(node as any).hash = "d3834b8c2656d347e13189469894be61";

export default node;
