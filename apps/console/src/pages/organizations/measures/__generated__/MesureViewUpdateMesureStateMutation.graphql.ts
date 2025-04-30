/**
 * @generated SignedSource<<8e8e8ce00d1ef971d4cb203c17d3abd6>>
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
    "cacheID": "59db8143f555230677a401f5bed6a6c3",
    "id": null,
    "metadata": {},
    "name": "MeasureViewUpdateMeasureStateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewUpdateMeasureStateMutation(\n  $input: UpdateMeasureInput!\n) {\n  updateMeasure(input: $input) {\n    measure {\n      id\n      state\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ae0f63f654568f052621f28e73382172";

export default node;
