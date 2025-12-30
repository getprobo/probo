/**
 * @generated SignedSource<<50c6ee9ebfd56ffeac153d8613118345>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteDatumInput = {
  datumId: string;
};
export type DatumGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDatumInput;
};
export type DatumGraphDeleteMutation$data = {
  readonly deleteDatum: {
    readonly deletedDatumId: string;
  };
};
export type DatumGraphDeleteMutation = {
  response: DatumGraphDeleteMutation$data;
  variables: DatumGraphDeleteMutation$variables;
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
  "name": "deletedDatumId",
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
    "name": "DatumGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDatumPayload",
        "kind": "LinkedField",
        "name": "deleteDatum",
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
    "name": "DatumGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDatumPayload",
        "kind": "LinkedField",
        "name": "deleteDatum",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedDatumId",
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
    "cacheID": "c8ac1dfe4d44edfa33008bfc92dee99d",
    "id": null,
    "metadata": {},
    "name": "DatumGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation DatumGraphDeleteMutation(\n  $input: DeleteDatumInput!\n) {\n  deleteDatum(input: $input) {\n    deletedDatumId\n  }\n}\n"
  }
};
})();

(node as any).hash = "1d6ffd466ce547e7377669b6ba567b9d";

export default node;
