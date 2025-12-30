/**
 * @generated SignedSource<<a1cf7f654fe2f662c446e60ce82e3ae9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateMeasureInput = {
  category: string;
  description?: string | null | undefined;
  name: string;
  organizationId: string;
};
export type MeasureFormDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMeasureInput;
};
export type MeasureFormDialogCreateMutation$data = {
  readonly createMeasure: {
    readonly measureEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"MeasureFormDialogMeasureFragment">;
      };
    };
  };
};
export type MeasureFormDialogCreateMutation = {
  response: MeasureFormDialogCreateMutation$data;
  variables: MeasureFormDialogCreateMutation$variables;
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
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeasurePayload",
        "kind": "LinkedField",
        "name": "createMeasure",
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
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "MeasureFormDialogMeasureFragment"
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
    "name": "MeasureFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeasurePayload",
        "kind": "LinkedField",
        "name": "createMeasure",
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
                    "name": "description",
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
                    "name": "category",
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
    "cacheID": "6ee48b7249e180aafe2f40c120ec1195",
    "id": null,
    "metadata": {},
    "name": "MeasureFormDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureFormDialogCreateMutation(\n  $input: CreateMeasureInput!\n) {\n  createMeasure(input: $input) {\n    measureEdge {\n      node {\n        ...MeasureFormDialogMeasureFragment\n        id\n      }\n    }\n  }\n}\n\nfragment MeasureFormDialogMeasureFragment on Measure {\n  id\n  description\n  name\n  category\n  state\n}\n"
  }
};
})();

(node as any).hash = "242f974b9926705919bf030281363dfc";

export default node;
