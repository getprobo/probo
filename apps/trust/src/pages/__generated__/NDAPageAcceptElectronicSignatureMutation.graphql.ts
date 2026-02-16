/**
 * @generated SignedSource<<41e415a03c63efbb590b1c2c782d5eac>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
export type AcceptElectronicSignatureInput = {
  fullName: string;
  signatureId: string;
};
export type NDAPageAcceptElectronicSignatureMutation$variables = {
  input: AcceptElectronicSignatureInput;
};
export type NDAPageAcceptElectronicSignatureMutation$data = {
  readonly acceptElectronicSignature: {
    readonly signature: {
      readonly id: string;
      readonly status: ElectronicSignatureStatus;
    };
  } | null | undefined;
};
export type NDAPageAcceptElectronicSignatureMutation = {
  response: NDAPageAcceptElectronicSignatureMutation$data;
  variables: NDAPageAcceptElectronicSignatureMutation$variables;
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
    "concreteType": "AcceptElectronicSignaturePayload",
    "kind": "LinkedField",
    "name": "acceptElectronicSignature",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ElectronicSignature",
        "kind": "LinkedField",
        "name": "signature",
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
            "name": "status",
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
    "name": "NDAPageAcceptElectronicSignatureMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NDAPageAcceptElectronicSignatureMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "db0c219a7133025b56252836d7f1a85d",
    "id": null,
    "metadata": {},
    "name": "NDAPageAcceptElectronicSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation NDAPageAcceptElectronicSignatureMutation(\n  $input: AcceptElectronicSignatureInput!\n) {\n  acceptElectronicSignature(input: $input) {\n    signature {\n      id\n      status\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b4ab543bd58878bb4968de6ea8b7c448";

export default node;
