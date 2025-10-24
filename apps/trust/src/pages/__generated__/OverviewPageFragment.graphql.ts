/**
 * @generated SignedSource<<cde79d7c929ba1be8e0d033c90e7eeb7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type CountryCode = "AD" | "AE" | "AF" | "AG" | "AI" | "AL" | "AM" | "AO" | "AQ" | "AR" | "AS" | "AT" | "AU" | "AW" | "AX" | "AZ" | "BA" | "BB" | "BD" | "BE" | "BF" | "BG" | "BH" | "BI" | "BJ" | "BL" | "BM" | "BN" | "BO" | "BQ" | "BR" | "BS" | "BT" | "BV" | "BW" | "BY" | "BZ" | "CA" | "CC" | "CD" | "CF" | "CG" | "CH" | "CI" | "CK" | "CL" | "CM" | "CN" | "CO" | "CR" | "CU" | "CV" | "CW" | "CX" | "CY" | "CZ" | "DE" | "DJ" | "DK" | "DM" | "DO" | "DZ" | "EC" | "EE" | "EG" | "EH" | "ER" | "ES" | "ET" | "EU" | "FI" | "FJ" | "FK" | "FM" | "FO" | "FR" | "GA" | "GB" | "GD" | "GE" | "GF" | "GG" | "GH" | "GI" | "GL" | "GM" | "GN" | "GP" | "GQ" | "GR" | "GT" | "GU" | "GW" | "GY" | "HK" | "HM" | "HN" | "HR" | "HT" | "HU" | "ID" | "IE" | "IL" | "IM" | "IN" | "IO" | "IQ" | "IR" | "IS" | "IT" | "JE" | "JM" | "JO" | "JP" | "KE" | "KG" | "KH" | "KI" | "KM" | "KN" | "KP" | "KR" | "KW" | "KY" | "KZ" | "LA" | "LB" | "LC" | "LI" | "LK" | "LR" | "LS" | "LT" | "LU" | "LV" | "LY" | "MA" | "MC" | "MD" | "ME" | "MF" | "MG" | "MH" | "MK" | "ML" | "MM" | "MN" | "MO" | "MP" | "MQ" | "MR" | "MS" | "MT" | "MU" | "MV" | "MW" | "MX" | "MY" | "MZ" | "NA" | "NC" | "NE" | "NF" | "NG" | "NI" | "NL" | "NO" | "NP" | "NR" | "NU" | "NZ" | "OM" | "PA" | "PE" | "PF" | "PG" | "PH" | "PK" | "PL" | "PM" | "PN" | "PR" | "PS" | "PT" | "PW" | "PY" | "QA" | "RE" | "RO" | "RS" | "RU" | "RW" | "SA" | "SB" | "SC" | "SD" | "SE" | "SG" | "SH" | "SI" | "SJ" | "SK" | "SL" | "SM" | "SN" | "SO" | "SR" | "SS" | "ST" | "SV" | "SX" | "SY" | "SZ" | "TC" | "TD" | "TF" | "TG" | "TH" | "TJ" | "TK" | "TL" | "TM" | "TN" | "TO" | "TR" | "TT" | "TV" | "TW" | "TZ" | "UA" | "UG" | "UM" | "US" | "UY" | "UZ" | "VA" | "VC" | "VE" | "VG" | "VI" | "VN" | "VU" | "WF" | "WS" | "YE" | "YT" | "ZA" | "ZM" | "ZW";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
import { FragmentRefs } from "relay-runtime";
export type OverviewPageFragment$data = {
  readonly documents: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly documentType: DocumentType;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentRowFragment">;
      };
    }>;
  };
  readonly references: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly logoUrl: string;
        readonly name: string;
        readonly websiteUrl: string;
      };
    }>;
  };
  readonly trustCenterFiles: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly category: string;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFileRowFragment">;
      };
    }>;
  };
  readonly vendors: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly countries: ReadonlyArray<CountryCode>;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"VendorRowFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "OverviewPageFragment";
};
export type OverviewPageFragment$key = {
  readonly " $data"?: OverviewPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OverviewPageFragment">;
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
  {
    "kind": "Literal",
    "name": "first",
    "value": 5
  }
];
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OverviewPageFragment",
  "selections": [
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 14
        }
      ],
      "concreteType": "TrustCenterReferenceConnection",
      "kind": "LinkedField",
      "name": "references",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenterReferenceEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "TrustCenterReference",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
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
                  "name": "logoUrl",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "websiteUrl",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "references(first:14)"
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 3
        }
      ],
      "concreteType": "VendorConnection",
      "kind": "LinkedField",
      "name": "vendors",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "VendorEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Vendor",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "countries",
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "VendorRowFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "vendors(first:3)"
    },
    {
      "alias": null,
      "args": (v1/*: any*/),
      "concreteType": "DocumentConnection",
      "kind": "LinkedField",
      "name": "documents",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "DocumentEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "Document",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "DocumentRowFragment"
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "documentType",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "documents(first:5)"
    },
    {
      "alias": null,
      "args": (v1/*: any*/),
      "concreteType": "TrustCenterFileConnection",
      "kind": "LinkedField",
      "name": "trustCenterFiles",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenterFileEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "TrustCenterFile",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v0/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "category",
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "TrustCenterFileRowFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "trustCenterFiles(first:5)"
    }
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "0f4e9ac02d068ba971bb1f5be5fdaccf";

export default node;
