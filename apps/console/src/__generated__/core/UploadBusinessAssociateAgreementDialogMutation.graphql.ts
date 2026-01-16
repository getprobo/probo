/**
 * @generated SignedSource<<f57c27c6042a9507adc37d4404b09b58>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadVendorBusinessAssociateAgreementInput = {
  file: any;
  fileName: string;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
  vendorId: string;
};
export type UploadBusinessAssociateAgreementDialogMutation$variables = {
  input: UploadVendorBusinessAssociateAgreementInput;
};
export type UploadBusinessAssociateAgreementDialogMutation$data = {
  readonly uploadVendorBusinessAssociateAgreement: {
    readonly vendorBusinessAssociateAgreement: {
      readonly createdAt: string;
      readonly fileName: string;
      readonly fileUrl: string;
      readonly id: string;
      readonly validFrom: string | null | undefined;
      readonly validUntil: string | null | undefined;
    };
  };
};
export type UploadBusinessAssociateAgreementDialogMutation = {
  response: UploadBusinessAssociateAgreementDialogMutation$data;
  variables: UploadBusinessAssociateAgreementDialogMutation$variables;
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
    "concreteType": "UploadVendorBusinessAssociateAgreementPayload",
    "kind": "LinkedField",
    "name": "uploadVendorBusinessAssociateAgreement",
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
            "name": "fileName",
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
    "name": "UploadBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UploadBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "bb7a0f67773da5b3916d08fdeaa2dc8d",
    "id": null,
    "metadata": {},
    "name": "UploadBusinessAssociateAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation UploadBusinessAssociateAgreementDialogMutation(\n  $input: UploadVendorBusinessAssociateAgreementInput!\n) {\n  uploadVendorBusinessAssociateAgreement(input: $input) {\n    vendorBusinessAssociateAgreement {\n      id\n      fileName\n      fileUrl\n      validFrom\n      validUntil\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e3fce209e1a4006a31996903e4c6cbae";

export default node;
