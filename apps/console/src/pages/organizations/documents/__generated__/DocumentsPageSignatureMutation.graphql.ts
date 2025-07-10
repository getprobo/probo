/**
 * @generated SignedSource<<4c3514470041046f80f410ea5a3275d6>>
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
export type DocumentsPageSignatureMutation$variables = {
  input: BulkRequestSignaturesInput;
};
export type DocumentsPageSignatureMutation$data = {
  readonly bulkRequestSignatures: {
    readonly documentVersionSignatureEdges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly state: DocumentVersionSignatureState;
      };
    }>;
  };
};
export type DocumentsPageSignatureMutation = {
  response: DocumentsPageSignatureMutation$data;
  variables: DocumentsPageSignatureMutation$variables;
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
    "name": "DocumentsPageSignatureMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentsPageSignatureMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a61a828bc67d8b569c38f4826b42d00f",
    "id": null,
    "metadata": {},
    "name": "DocumentsPageSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentsPageSignatureMutation(\n  $input: BulkRequestSignaturesInput!\n) {\n  bulkRequestSignatures(input: $input) {\n    documentVersionSignatureEdges {\n      node {\n        id\n        state\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1cb2b7dfa938b35cf75b09b51cdd2dca";

export default node;
