/**
 * @generated SignedSource<<a26f31885b4f92b2ad9398bce7f538cf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type BulkDeleteDocumentsInput = {
  documentIds: ReadonlyArray<string>;
};
export type DocumentGraphBulkDeleteDocumentsMutation$variables = {
  input: BulkDeleteDocumentsInput;
};
export type DocumentGraphBulkDeleteDocumentsMutation$data = {
  readonly bulkDeleteDocuments: {
    readonly deletedDocumentIds: ReadonlyArray<string>;
  };
};
export type DocumentGraphBulkDeleteDocumentsMutation = {
  response: DocumentGraphBulkDeleteDocumentsMutation$data;
  variables: DocumentGraphBulkDeleteDocumentsMutation$variables;
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
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedDocumentIds",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentGraphBulkDeleteDocumentsMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "BulkDeleteDocumentsPayload",
        "kind": "LinkedField",
        "name": "bulkDeleteDocuments",
        "plural": false,
        "selections": [
          (v2/*: any*/)
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
    "name": "DocumentGraphBulkDeleteDocumentsMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "BulkDeleteDocumentsPayload",
        "kind": "LinkedField",
        "name": "bulkDeleteDocuments",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteRecord",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedDocumentIds"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "655f43d282b8e0007f3b93e381e7ff84",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphBulkDeleteDocumentsMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentGraphBulkDeleteDocumentsMutation(\n  $input: BulkDeleteDocumentsInput!\n) {\n  bulkDeleteDocuments(input: $input) {\n    deletedDocumentIds\n  }\n}\n"
  }
};
})();

(node as any).hash = "919857a24f366b8ebe004fcadf503c3d";

export default node;
