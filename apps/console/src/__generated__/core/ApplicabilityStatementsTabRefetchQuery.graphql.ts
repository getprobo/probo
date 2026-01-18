/**
 * @generated SignedSource<<601895cf055c15f58d472561a1666634>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ApplicabilityStatementsTabRefetchQuery$variables = {
  id: string;
};
export type ApplicabilityStatementsTabRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"ApplicabilityStatementsTabFragment">;
  };
};
export type ApplicabilityStatementsTabRefetchQuery = {
  response: ApplicabilityStatementsTabRefetchQuery$data;
  variables: ApplicabilityStatementsTabRefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "id"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ApplicabilityStatementsTabRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "ApplicabilityStatementsTabFragment"
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ApplicabilityStatementsTabRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": "applicabilityStatementsInfo",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "first",
                    "value": 0
                  }
                ],
                "concreteType": "ApplicabilityStatementConnection",
                "kind": "LinkedField",
                "name": "applicabilityStatements",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "totalCount",
                    "storageKey": null
                  }
                ],
                "storageKey": "applicabilityStatements(first:0)"
              },
              {
                "alias": "canCreateApplicabilityStatement",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:state-of-applicability-control-mapping:create"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:create\")"
              },
              {
                "alias": "canDeleteApplicabilityStatement",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:state-of-applicability-control-mapping:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:delete\")"
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "AvailableStateOfApplicabilityControl",
                "kind": "LinkedField",
                "name": "availableControls",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "controlId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "sectionTitle",
                    "storageKey": null
                  },
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
                    "name": "frameworkId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "frameworkName",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "organizationId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "applicabilityStatementId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "stateOfApplicabilityId",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "applicability",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "justification",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "bestPractice",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "regulatory",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "contractual",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "riskAssessment",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "StateOfApplicability",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "2b4d635073c8471e0f642aeec7c84a0b",
    "id": null,
    "metadata": {},
    "name": "ApplicabilityStatementsTabRefetchQuery",
    "operationKind": "query",
    "text": "query ApplicabilityStatementsTabRefetchQuery(\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...ApplicabilityStatementsTabFragment\n    id\n  }\n}\n\nfragment ApplicabilityStatementsTabFragment on StateOfApplicability {\n  id\n  applicabilityStatementsInfo: applicabilityStatements(first: 0) {\n    totalCount\n  }\n  canCreateApplicabilityStatement: permission(action: \"core:state-of-applicability-control-mapping:create\")\n  canDeleteApplicabilityStatement: permission(action: \"core:state-of-applicability-control-mapping:delete\")\n  availableControls {\n    controlId\n    sectionTitle\n    name\n    frameworkId\n    frameworkName\n    organizationId\n    applicabilityStatementId\n    stateOfApplicabilityId\n    applicability\n    justification\n    bestPractice\n    regulatory\n    contractual\n    riskAssessment\n  }\n}\n"
  }
};
})();

(node as any).hash = "8183adc2185e14fe2962c58058bef8ed";

export default node;
