/**
 * @generated SignedSource<<24e3fd39e535ba08f8099a305fcba487>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlReportMappingInput = {
  controlId: string;
  reportId: string;
};
export type FrameworkControlPageAttachReportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlReportMappingInput;
};
export type FrameworkControlPageAttachReportMutation$data = {
  readonly createControlReportMapping: {
    readonly reportEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedReportsCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachReportMutation = {
  response: FrameworkControlPageAttachReportMutation$data;
  variables: FrameworkControlPageAttachReportMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkControlPageAttachReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlReportMappingPayload",
        "kind": "LinkedField",
        "name": "createControlReportMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ReportEdge",
            "kind": "LinkedField",
            "name": "reportEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Report",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedReportsCardFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "FrameworkControlPageAttachReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlReportMappingPayload",
        "kind": "LinkedField",
        "name": "createControlReportMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ReportEdge",
            "kind": "LinkedField",
            "name": "reportEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Report",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
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
                    "name": "state",
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
                    "concreteType": "Framework",
                    "kind": "LinkedField",
                    "name": "framework",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "reportEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f2488599c5b98e13c281669d4c4b7224",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachReportMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachReportMutation(\n  $input: CreateControlReportMappingInput!\n) {\n  createControlReportMapping(input: $input) {\n    reportEdge {\n      node {\n        id\n        ...LinkedReportsCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedReportsCardFragment on Report {\n  id\n  name\n  createdAt\n  state\n  validFrom\n  validUntil\n  framework {\n    id\n    name\n  }\n}\n"
  }
};
})();

(node as any).hash = "835161daae59fc5bd33af1cbcdaec9f0";

export default node;
