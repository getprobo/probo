/**
 * @generated SignedSource<<4c1c57380bf4ac13f476346c8bee05fe>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateMeasureInput = {
  category: string;
  description: string;
  name: string;
  organizationId: string;
};
export type NewMeasureViewCreateMeasureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMeasureInput;
};
export type NewMeasureViewCreateMeasureMutation$data = {
  readonly createMeasure: {
    readonly measureEdge: {
      readonly node: {
        readonly description: string;
        readonly id: string;
        readonly name: string;
      };
    };
  };
};
export type NewMeasureViewCreateMeasureMutation = {
  response: NewMeasureViewCreateMeasureMutation$data;
  variables: NewMeasureViewCreateMeasureMutation$variables;
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
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "id",
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
          "name": "description",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "NewMeasureViewCreateMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeasurePayload",
        "kind": "LinkedField",
        "name": "createMeasure",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "NewMeasureViewCreateMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeasurePayload",
        "kind": "LinkedField",
        "name": "createMeasure",
        "plural": false,
        "selections": [
          (v3/*: any*/),
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
    "cacheID": "de8b930c9628c62ae4272bd427076ebd",
    "id": null,
    "metadata": {},
    "name": "NewMeasureViewCreateMeasureMutation",
    "operationKind": "mutation",
    "text": "mutation NewMeasureViewCreateMeasureMutation(\n  $input: CreateMeasureInput!\n) {\n  createMeasure(input: $input) {\n    measureEdge {\n      node {\n        id\n        name\n        description\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "94532353875a658b6e0b87a8181eab5e";

export default node;
