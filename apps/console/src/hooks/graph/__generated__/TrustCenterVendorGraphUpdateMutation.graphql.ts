/**
 * @generated SignedSource<<012c4e605372a38efef4780c63d40888>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
export type UpdateVendorInput = {
  businessAssociateAgreementUrl?: string | null | undefined;
  businessOwnerId?: string | null | undefined;
  category?: VendorCategory | null | undefined;
  certifications?: ReadonlyArray<string> | null | undefined;
  dataProcessingAgreementUrl?: string | null | undefined;
  description?: string | null | undefined;
  headquarterAddress?: string | null | undefined;
  id: string;
  legalName?: string | null | undefined;
  name?: string | null | undefined;
  privacyPolicyUrl?: string | null | undefined;
  securityOwnerId?: string | null | undefined;
  securityPageUrl?: string | null | undefined;
  serviceLevelAgreementUrl?: string | null | undefined;
  showOnTrustCenter?: boolean | null | undefined;
  statusPageUrl?: string | null | undefined;
  subprocessorsListUrl?: string | null | undefined;
  termsOfServiceUrl?: string | null | undefined;
  trustPageUrl?: string | null | undefined;
  websiteUrl?: string | null | undefined;
};
export type TrustCenterVendorGraphUpdateMutation$variables = {
  input: UpdateVendorInput;
};
export type TrustCenterVendorGraphUpdateMutation$data = {
  readonly updateVendor: {
    readonly vendor: {
      readonly id: string;
      readonly showOnTrustCenter: boolean;
    };
  };
};
export type TrustCenterVendorGraphUpdateMutation = {
  response: TrustCenterVendorGraphUpdateMutation$data;
  variables: TrustCenterVendorGraphUpdateMutation$variables;
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
    "concreteType": "UpdateVendorPayload",
    "kind": "LinkedField",
    "name": "updateVendor",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Vendor",
        "kind": "LinkedField",
        "name": "vendor",
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
            "name": "showOnTrustCenter",
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
    "name": "TrustCenterVendorGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterVendorGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "44e483af96833ea9efb1567fa5983282",
    "id": null,
    "metadata": {},
    "name": "TrustCenterVendorGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterVendorGraphUpdateMutation(\n  $input: UpdateVendorInput!\n) {\n  updateVendor(input: $input) {\n    vendor {\n      id\n      showOnTrustCenter\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e561918e67c08f1e3fae85dfb04591d1";

export default node;
