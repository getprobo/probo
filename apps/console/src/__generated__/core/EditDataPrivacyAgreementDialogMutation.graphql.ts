/**
 * @generated SignedSource<<12863d17356b3d0f16a9c377cd95eb5b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateVendorDataPrivacyAgreementInput = {
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
  vendorId: string;
};
export type EditDataPrivacyAgreementDialogMutation$variables = {
  input: UpdateVendorDataPrivacyAgreementInput;
};
export type EditDataPrivacyAgreementDialogMutation$data = {
  readonly updateVendorDataPrivacyAgreement: {
    readonly vendorDataPrivacyAgreement: {
      readonly createdAt: string;
      readonly fileUrl: string;
      readonly id: string;
      readonly validFrom: string | null | undefined;
      readonly validUntil: string | null | undefined;
    };
  };
};
export type EditDataPrivacyAgreementDialogMutation = {
  response: EditDataPrivacyAgreementDialogMutation$data;
  variables: EditDataPrivacyAgreementDialogMutation$variables;
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
    "concreteType": "UpdateVendorDataPrivacyAgreementPayload",
    "kind": "LinkedField",
    "name": "updateVendorDataPrivacyAgreement",
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
    "name": "EditDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e84e9f9457d6e4dc295578e540aa3e64",
    "id": null,
    "metadata": {},
    "name": "EditDataPrivacyAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditDataPrivacyAgreementDialogMutation(\n  $input: UpdateVendorDataPrivacyAgreementInput!\n) {\n  updateVendorDataPrivacyAgreement(input: $input) {\n    vendorDataPrivacyAgreement {\n      id\n      fileUrl\n      validFrom\n      validUntil\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b2caa3bc89953896a848ee77fd96a767";

export default node;
