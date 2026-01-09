/**
 * @generated SignedSource<<ea70fd099349ca45cb2a847b06e28694>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
export type SignDocumentInput = {
  documentVersionId: string;
};
export type EmployeeDocumentSignaturePageSignMutation$variables = {
  input: SignDocumentInput;
};
export type EmployeeDocumentSignaturePageSignMutation$data = {
  readonly signDocument: {
    readonly documentVersionSignature: {
      readonly id: string;
      readonly state: DocumentVersionSignatureState;
    };
  };
};
export type EmployeeDocumentSignaturePageSignMutation = {
  response: EmployeeDocumentSignaturePageSignMutation$data;
  variables: EmployeeDocumentSignaturePageSignMutation$variables;
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
    "concreteType": "SignDocumentPayload",
    "kind": "LinkedField",
    "name": "signDocument",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "DocumentVersionSignature",
        "kind": "LinkedField",
        "name": "documentVersionSignature",
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
    "name": "EmployeeDocumentSignaturePageSignMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EmployeeDocumentSignaturePageSignMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "dfdd778fc4b0cfc9c007e4c29258ea96",
    "id": null,
    "metadata": {},
    "name": "EmployeeDocumentSignaturePageSignMutation",
    "operationKind": "mutation",
    "text": "mutation EmployeeDocumentSignaturePageSignMutation(\n  $input: SignDocumentInput!\n) {\n  signDocument(input: $input) {\n    documentVersionSignature {\n      id\n      state\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b91674332a1914e270e4fb811ccfd479";

export default node;
