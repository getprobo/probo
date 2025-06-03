/**
 * @generated SignedSource<<68ddff15cd65f5367f7e9ae6b3c38f34>>
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
  readonly assessedAt: any;
  readonly assessedBy: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly businessImpact: BusinessImpact;
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

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorRiskAssessmentTabFragment_assessment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "assessedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "assessedBy",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
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
})();

(node as any).hash = "08d6de198077f9b5985258466bbc33c0";

export default node;
