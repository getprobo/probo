/**
 * @generated SignedSource<<baa13bd4048002970217bfb1e8017e05>>
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
export type EditMeasureViewUpdateMeasureMutation$variables = {
  input: UpdateMeasureInput;
};
export type EditMeasureViewUpdateMeasureMutation$data = {
  readonly updateMeasure: {
    readonly measure: {
      readonly description: string;
      readonly id: string;
      readonly name: string;
    };
  };
};
export type EditMeasureViewUpdateMeasureMutation = {
  response: EditMeasureViewUpdateMeasureMutation$data;
  variables: EditMeasureViewUpdateMeasureMutation$variables;
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
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EditMeasureViewUpdateMeasureMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditMeasureViewUpdateMeasureMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0d6bab0c03a2c72635e1e42fe69c9551",
    "id": null,
    "metadata": {},
    "name": "EditMeasureViewUpdateMeasureMutation",
    "operationKind": "mutation",
    "text": "mutation EditMeasureViewUpdateMeasureMutation(\n  $input: UpdateMeasureInput!\n) {\n  updateMeasure(input: $input) {\n    measure {\n      id\n      name\n      description\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1b91ac6890f48a7f536adcba303f1718";

export default node;
