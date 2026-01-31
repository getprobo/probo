/**
 * @generated SignedSource<<b0dc20be7382cf67a0906b4f06e713e0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateTrustCenterBrandInput = {
  darkLogoFile?: any | null | undefined;
  logoFile?: any | null | undefined;
  trustCenterId: string;
};
export type CompliancePageBrandPage_updateMutation$variables = {
  input: UpdateTrustCenterBrandInput;
};
export type CompliancePageBrandPage_updateMutation$data = {
  readonly updateTrustCenterBrand: {
    readonly trustCenter: {
      readonly darkLogoFileUrl: string | null | undefined;
      readonly id: string;
      readonly logoFileUrl: string | null | undefined;
    };
  };
};
export type CompliancePageBrandPage_updateMutation = {
  response: CompliancePageBrandPage_updateMutation$data;
  variables: CompliancePageBrandPage_updateMutation$variables;
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
    "concreteType": "UpdateTrustCenterBrandPayload",
    "kind": "LinkedField",
    "name": "updateTrustCenterBrand",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenter",
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
            "name": "logoFileUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "darkLogoFileUrl",
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
    "name": "CompliancePageBrandPage_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageBrandPage_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fb398aa8a46a93ca7b68e4ef9841a4d6",
    "id": null,
    "metadata": {},
    "name": "CompliancePageBrandPage_updateMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageBrandPage_updateMutation(\n  $input: UpdateTrustCenterBrandInput!\n) {\n  updateTrustCenterBrand(input: $input) {\n    trustCenter {\n      id\n      logoFileUrl\n      darkLogoFileUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "cb158c24dbd3824e6fa9cc3e4632dc81";

export default node;
