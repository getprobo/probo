/**
 * @generated SignedSource<<26343418dbdd4baf63f2846470e2d3ab>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorBusinessAssociateAgreementInput = {
  vendorId: string;
};
export type DeleteBusinessAssociateAgreementDialogMutation$variables = {
  input: DeleteVendorBusinessAssociateAgreementInput;
};
export type DeleteBusinessAssociateAgreementDialogMutation$data = {
  readonly deleteVendorBusinessAssociateAgreement: {
    readonly deletedVendorId: string;
  };
};
export type DeleteBusinessAssociateAgreementDialogMutation = {
  response: DeleteBusinessAssociateAgreementDialogMutation$data;
  variables: DeleteBusinessAssociateAgreementDialogMutation$variables;
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
    "concreteType": "DeleteVendorBusinessAssociateAgreementPayload",
    "kind": "LinkedField",
    "name": "deleteVendorBusinessAssociateAgreement",
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
    "name": "DeleteBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DeleteBusinessAssociateAgreementDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c8aa903ee410ee9e591c82f7f75e044c",
    "id": null,
    "metadata": {},
    "name": "DeleteBusinessAssociateAgreementDialogMutation",
    "operationKind": "mutation",
    "text": "mutation DeleteBusinessAssociateAgreementDialogMutation(\n  $input: DeleteVendorBusinessAssociateAgreementInput!\n) {\n  deleteVendorBusinessAssociateAgreement(input: $input) {\n    deletedVendorId\n  }\n}\n"
  }
};
})();

(node as any).hash = "39106d060e4e1da4eaa78a8205fe7cf8";

export default node;
