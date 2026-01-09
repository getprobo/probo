/**
 * @generated SignedSource<<f114dd31b601c0097380450ab5029b01>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteDraftDocumentVersionInput = {
  documentVersionId: string;
};
export type DocumentGraphDeleteDraftMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteDraftDocumentVersionInput;
};
export type DocumentGraphDeleteDraftMutation$data = {
  readonly deleteDraftDocumentVersion: {
    readonly deletedDocumentVersionId: string;
  };
};
export type DocumentGraphDeleteDraftMutation = {
  response: DocumentGraphDeleteDraftMutation$data;
  variables: DocumentGraphDeleteDraftMutation$variables;
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
  "name": "deletedDocumentVersionId",
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
    "name": "DocumentGraphDeleteDraftMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDraftDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "deleteDraftDocumentVersion",
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
    "name": "DocumentGraphDeleteDraftMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteDraftDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "deleteDraftDocumentVersion",
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
            "name": "deletedDocumentVersionId",
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
    "cacheID": "cd7e8253362e1d7e719e4ae2c206756e",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphDeleteDraftMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentGraphDeleteDraftMutation(\n  $input: DeleteDraftDocumentVersionInput!\n) {\n  deleteDraftDocumentVersion(input: $input) {\n    deletedDocumentVersionId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d43940b41e3838fd03c3341c13b4c5c5";

export default node;
