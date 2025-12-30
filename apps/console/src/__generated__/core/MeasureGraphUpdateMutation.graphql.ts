/**
 * @generated SignedSource<<0568c40c0caa32ce6daf11d414e2cc72>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type UpdateMeasureInput = {
  category?: string | null | undefined;
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  state?: MeasureState | null | undefined;
};
export type MeasureGraphUpdateMutation$variables = {
  input: UpdateMeasureInput;
};
export type MeasureGraphUpdateMutation$data = {
  readonly updateMeasure: {
    readonly measure: {
      readonly " $fragmentSpreads": FragmentRefs<"MeasureFormDialogMeasureFragment">;
    };
  };
};
export type MeasureGraphUpdateMutation = {
  response: MeasureGraphUpdateMutation$data;
  variables: MeasureGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureGraphUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateMeasurePayload",
        "kind": "LinkedField",
        "name": "updateMeasure",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Measure",
            "kind": "LinkedField",
            "name": "measure",
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
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureGraphUpdateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateMeasurePayload",
        "kind": "LinkedField",
        "name": "updateMeasure",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Measure",
            "kind": "LinkedField",
            "name": "measure",
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
      }
    ]
  },
  "params": {
    "cacheID": "17d35db59a42796407f7f0d6bb4798ba",
    "id": null,
    "metadata": {},
    "name": "MeasureGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureGraphUpdateMutation(\n  $input: UpdateMeasureInput!\n) {\n  updateMeasure(input: $input) {\n    measure {\n      ...MeasureFormDialogMeasureFragment\n      id\n    }\n  }\n}\n\nfragment MeasureFormDialogMeasureFragment on Measure {\n  id\n  description\n  name\n  category\n  state\n}\n"
  }
};
})();

(node as any).hash = "9f74c8c6cea82050d077d26e6b7547f6";

export default node;
