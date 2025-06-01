/**
 * @generated SignedSource<<ea3e7fd6a3b7dd5f78622e9f21cdb1d3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type useVendorFormFragment$data = {
  readonly businessOwner: {
    readonly id: string;
  } | null | undefined;
  readonly certifications: ReadonlyArray<string>;
  readonly dataProcessingAgreementUrl: string | null | undefined;
  readonly description: string | null | undefined;
  readonly headquarterAddress: string | null | undefined;
  readonly id: string;
  readonly legalName: string | null | undefined;
  readonly name: string;
  readonly privacyPolicyUrl: string | null | undefined;
  readonly securityOwner: {
    readonly id: string;
  } | null | undefined;
  readonly securityPageUrl: string | null | undefined;
  readonly serviceLevelAgreementUrl: string | null | undefined;
  readonly statusPageUrl: string | null | undefined;
  readonly termsOfServiceUrl: string | null | undefined;
  readonly trustPageUrl: string | null | undefined;
  readonly websiteUrl: string | null | undefined;
  readonly " $fragmentType": "useVendorFormFragment";
};
export type useVendorFormFragment$key = {
  readonly " $data"?: useVendorFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"useVendorFormFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = [
  (v0/*: any*/)
];
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "useVendorFormFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "description",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "statusPageUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "termsOfServiceUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "privacyPolicyUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "serviceLevelAgreementUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "dataProcessingAgreementUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "websiteUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "legalName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "headquarterAddress",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "certifications",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "securityPageUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "trustPageUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "businessOwner",
      "plural": false,
      "selections": (v1/*: any*/),
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "securityOwner",
      "plural": false,
      "selections": (v1/*: any*/),
      "storageKey": null
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};
})();

(node as any).hash = "87f1029f8c634a5f7efdb6d7abe17709";

export default node;
