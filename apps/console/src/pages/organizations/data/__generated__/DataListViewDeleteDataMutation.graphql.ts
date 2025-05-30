/**
 * @generated SignedSource<<0957856377454a1d927ba7bcf1d84cac>>
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
export type DataListViewDeleteDataMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDatumInput;
};
export type DataListViewDeleteDataMutation$data = {
  readonly deleteDatum: {
    readonly deletedDatumId: string;
  };
};
export type DataListViewDeleteDataMutation = {
  response: DataListViewDeleteDataMutation$data;
  variables: DataListViewDeleteDataMutation$variables;
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
    "name": "DataListViewDeleteDataMutation",
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
    "name": "DataListViewDeleteDataMutation",
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
    "cacheID": "27043dcacab7e2b4be1fb792f26dafff",
    "id": null,
    "metadata": {},
    "name": "DataListViewDeleteDataMutation",
    "operationKind": "mutation",
    "text": "mutation DataListViewDeleteDataMutation(\n  $input: DeleteDatumInput!\n) {\n  deleteDatum(input: $input) {\n    deletedDatumId\n  }\n}\n"
  }
};
})();

(node as any).hash = "05293e1e936af583e13073c8c06d14dd";

export default node;
