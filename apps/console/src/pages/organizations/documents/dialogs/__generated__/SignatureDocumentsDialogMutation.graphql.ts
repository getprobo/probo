/**
 * @generated SignedSource<<14890a4f10a8862a8aaf30ee1dc6f065>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type BulkRequestSignaturesInput = {
  documentIds: ReadonlyArray<string>;
  signatoryIds: ReadonlyArray<string>;
};
export type SignatureDocumentsDialogMutation$variables = {
  input: BulkRequestSignaturesInput;
};
export type SignatureDocumentsDialogMutation$data = {
  readonly bulkRequestSignatures: {
    readonly documentVersionSignatureEdges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly state: DocumentVersionSignatureState;
      };
    }>;
  };
};
export type SignatureDocumentsDialogMutation = {
  response: SignatureDocumentsDialogMutation$data;
  variables: SignatureDocumentsDialogMutation$variables;
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
    "concreteType": "BulkRequestSignaturesPayload",
    "kind": "LinkedField",
    "name": "bulkRequestSignatures",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "DocumentVersionSignatureEdge",
        "kind": "LinkedField",
        "name": "documentVersionSignatureEdges",
        "plural": true,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentVersionSignature",
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
                "name": "state",
                "storageKey": null
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SignatureDocumentsDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SignatureDocumentsDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "4426d76c855d441caab5c932c2ceb2ba",
    "id": null,
    "metadata": {},
    "name": "SignatureDocumentsDialogMutation",
    "operationKind": "mutation",
    "text": "mutation SignatureDocumentsDialogMutation(\n  $input: BulkRequestSignaturesInput!\n) {\n  bulkRequestSignatures(input: $input) {\n    documentVersionSignatureEdges {\n      node {\n        id\n        state\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1dbd9a7870416bf75d0a6fe44ebde408";

export default node;
