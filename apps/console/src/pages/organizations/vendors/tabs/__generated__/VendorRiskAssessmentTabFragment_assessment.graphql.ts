/**
 * @generated SignedSource<<b6963cdf0e4c8a5c6bb293cb53f8d404>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type BusinessImpact = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM";
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
import { FragmentRefs } from "relay-runtime";
export type VendorRiskAssessmentTabFragment_assessment$data = {
  readonly businessImpact: BusinessImpact;
  readonly createdAt: any;
  readonly dataSensitivity: DataSensitivity;
  readonly expiresAt: any;
  readonly id: string;
  readonly notes: string | null | undefined;
  readonly " $fragmentType": "VendorRiskAssessmentTabFragment_assessment";
};
export type VendorRiskAssessmentTabFragment_assessment$key = {
  readonly " $data"?: VendorRiskAssessmentTabFragment_assessment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorRiskAssessmentTabFragment_assessment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorRiskAssessmentTabFragment_assessment",
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
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "expiresAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "dataSensitivity",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "businessImpact",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "notes",
      "storageKey": null
    }
  ],
  "type": "VendorRiskAssessment",
  "abstractKey": null
};

(node as any).hash = "533771f5fac9e732556929fa2b1d0dc5";

export default node;
