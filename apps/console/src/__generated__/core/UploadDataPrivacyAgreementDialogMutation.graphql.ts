/**
 * @generated SignedSource<<0e9455daba78b038455718505f53a4d7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadVendorDataPrivacyAgreementInput = {
  file: any;
  fileName: string;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
  vendorId: string;
};
export type UploadDataPrivacyAgreementDialogMutation$variables = {
  input: UploadVendorDataPrivacyAgreementInput;
};
export type UploadDataPrivacyAgreementDialogMutation$data = {
  readonly uploadVendorDataPrivacyAgreement: {
    readonly vendorDataPrivacyAgreement: {
      readonly createdAt: string;
      readonly fileName: string;
      readonly fileUrl: string;
      readonly id: string;
      readonly validFrom: string | null | undefined;
      readonly validUntil: string | null | undefined;
    };
  };
};
export type UploadDataPrivacyAgreementDialogMutation = {
  response: UploadDataPrivacyAgreementDialogMutation$data;
  variables: UploadDataPrivacyAgreementDialogMutation$variables;
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
    "concreteType": "UploadVendorDataPrivacyAgreementPayload",
    "kind": "LinkedField",
    "name": "uploadVendorDataPrivacyAgreement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "VendorDataPrivacyAgreement",
        "kind": "LinkedField",
        "name": "vendorDataPrivacyAgreement",
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
    "name": "UploadDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UploadDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c59302d882fc9406dd7f09072db552b5",
    "id": null,
    "metadata": {},
    "name": "UploadDataPrivacyAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation UploadDataPrivacyAgreementDialogMutation(\n  $input: UploadVendorDataPrivacyAgreementInput!\n) {\n  uploadVendorDataPrivacyAgreement(input: $input) {\n    vendorDataPrivacyAgreement {\n      id\n      fileName\n      fileUrl\n      validFrom\n      validUntil\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8e4e70e111836cebd850061a75872980";

export default node;
