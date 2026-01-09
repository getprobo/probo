/**
 * @generated SignedSource<<4d1711d74f82bc8952b162a37dd8931e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorDataPrivacyAgreementInput = {
  vendorId: string;
};
export type DeleteDataPrivacyAgreementDialogMutation$variables = {
  input: DeleteVendorDataPrivacyAgreementInput;
};
export type DeleteDataPrivacyAgreementDialogMutation$data = {
  readonly deleteVendorDataPrivacyAgreement: {
    readonly deletedVendorId: string;
  };
};
export type DeleteDataPrivacyAgreementDialogMutation = {
  response: DeleteDataPrivacyAgreementDialogMutation$data;
  variables: DeleteDataPrivacyAgreementDialogMutation$variables;
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
    "concreteType": "DeleteVendorDataPrivacyAgreementPayload",
    "kind": "LinkedField",
    "name": "deleteVendorDataPrivacyAgreement",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "deletedVendorId",
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
    "name": "DeleteDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DeleteDataPrivacyAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "683fac4979ff59534ce2acaa05d82063",
    "id": null,
    "metadata": {},
    "name": "DeleteDataPrivacyAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation DeleteDataPrivacyAgreementDialogMutation(\n  $input: DeleteVendorDataPrivacyAgreementInput!\n) {\n  deleteVendorDataPrivacyAgreement(input: $input) {\n    deletedVendorId\n  }\n}\n"
  }
};
})();

(node as any).hash = "38cb9bcb710db551eda7ed8cff9fca5a";

export default node;
