/**
 * @generated SignedSource<<8d5961820ccfa705a899582a9a8ca953>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateVendorBusinessAssociateAgreementInput = {
  validFrom?: any | null | undefined;
  validUntil?: any | null | undefined;
  vendorId: string;
};
export type EditBusinessAssociateAgreementDialogMutation$variables = {
  input: UpdateVendorBusinessAssociateAgreementInput;
};
export type EditBusinessAssociateAgreementDialogMutation$data = {
  readonly updateVendorBusinessAssociateAgreement: {
    readonly vendorBusinessAssociateAgreement: {
      readonly createdAt: any;
      readonly fileUrl: string;
      readonly id: string;
      readonly validFrom: any | null | undefined;
      readonly validUntil: any | null | undefined;
    };
  };
};
export type EditBusinessAssociateAgreementDialogMutation = {
  response: EditBusinessAssociateAgreementDialogMutation$data;
  variables: EditBusinessAssociateAgreementDialogMutation$variables;
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
    "concreteType": "UpdateVendorBusinessAssociateAgreementPayload",
    "kind": "LinkedField",
    "name": "updateVendorBusinessAssociateAgreement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "VendorBusinessAssociateAgreement",
        "kind": "LinkedField",
        "name": "vendorBusinessAssociateAgreement",
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
            "name": "fileUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "validFrom",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "validUntil",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
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
    "name": "EditBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2e50e54d255a9bea661d82188913ed9c",
    "id": null,
    "metadata": {},
    "name": "EditBusinessAssociateAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditBusinessAssociateAgreementDialogMutation(\n  $input: UpdateVendorBusinessAssociateAgreementInput!\n) {\n  updateVendorBusinessAssociateAgreement(input: $input) {\n    vendorBusinessAssociateAgreement {\n      id\n      fileUrl\n      validFrom\n      validUntil\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "10b614dbbdee430ed650b94051be4bbc";

export default node;
