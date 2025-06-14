/**
 * @generated SignedSource<<1d23c8e1a8e16a9811d75d208ade0889>>
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
export type FrameworkControlPageAttachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlMeasureMappingInput;
};
export type FrameworkControlPageAttachMutation$data = {
  readonly createControlMeasureMapping: {
    readonly measureEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedMeasuresCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachMutation = {
  response: FrameworkControlPageAttachMutation$data;
  variables: FrameworkControlPageAttachMutation$variables;
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
    "name": "FrameworkControlPageAttachMutation",
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
    "name": "FrameworkControlPageAttachMutation",
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
    "cacheID": "85f870b18745fcd28bac3a52c404e98f",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachMutation(\n  $input: CreateControlMeasureMappingInput!\n) {\n  createControlMeasureMapping(input: $input) {\n    measureEdge {\n      node {\n        id\n        ...LinkedMeasuresCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n"
  }
};
})();

(node as any).hash = "e52f26e6906aa014ad020bd6f23a75d6";

export default node;
