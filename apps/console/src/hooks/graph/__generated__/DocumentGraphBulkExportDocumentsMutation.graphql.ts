/**
 * @generated SignedSource<<9b84adc0253186c6f0a5f47837d83d7f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type BulkExportDocumentsInput = {
  documentIds: ReadonlyArray<string>;
  watermarkEmail?: string | null | undefined;
  withSignatures: boolean;
  withWatermark: boolean;
};
export type DocumentGraphBulkExportDocumentsMutation$variables = {
  input: BulkExportDocumentsInput;
};
export type DocumentGraphBulkExportDocumentsMutation$data = {
  readonly bulkExportDocuments: {
    readonly exportJobId: string;
  };
};
export type DocumentGraphBulkExportDocumentsMutation = {
  response: DocumentGraphBulkExportDocumentsMutation$data;
  variables: DocumentGraphBulkExportDocumentsMutation$variables;
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
    "concreteType": "BulkExportDocumentsPayload",
    "kind": "LinkedField",
    "name": "bulkExportDocuments",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "exportJobId",
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
    "name": "DocumentGraphBulkExportDocumentsMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentGraphBulkExportDocumentsMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2f4eb206613f8fba307aeee5bad27b93",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphBulkExportDocumentsMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentGraphBulkExportDocumentsMutation(\n  $input: BulkExportDocumentsInput!\n) {\n  bulkExportDocuments(input: $input) {\n    exportJobId\n  }\n}\n"
  }
};
})();

(node as any).hash = "0a7bae4b58713bcc824a774c31bd6dca";

export default node;
