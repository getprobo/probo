/**
 * @generated SignedSource<<5c3c644d4844322b34c5fbee967880a8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type UpdateMeasureInput = {
  category?: string | null | undefined;
  description?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  state?: MeasureState | null | undefined;
};
export type MeasureViewUpdateMeasureStateMutation$variables = {
  input: UpdateMeasureInput;
};
export type MeasureViewUpdateMeasureStateMutation$data = {
  readonly updateMeasure: {
    readonly measure: {
      readonly id: string;
      readonly state: MeasureState;
    };
  };
};
export type MeasureViewUpdateMeasureStateMutation = {
  response: MeasureViewUpdateMeasureStateMutation$data;
  variables: MeasureViewUpdateMeasureStateMutation$variables;
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
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
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
            "name": "state",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureViewUpdateMeasureStateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewUpdateMeasureStateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0b67944fab3e9ae13f437aeee2b28c5b",
    "id": null,
    "metadata": {},
    "name": "MeasureViewUpdateMeasureStateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewUpdateMeasureStateMutation(\n  $input: UpdateMeasureInput!\n) {\n  updateMeasure(input: $input) {\n    measure {\n      id\n      state\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "d89e0bc25930592a1f723bbbdb3c43b8";

export default node;
