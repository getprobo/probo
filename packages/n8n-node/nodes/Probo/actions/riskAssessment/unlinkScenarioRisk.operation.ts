// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Scenario ID',
		name: 'riskAssessmentScenarioId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['unlinkScenarioRisk'],
			},
		},
		default: '',
		description: 'The ID of the scenario',
		required: true,
	},
	{
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['unlinkScenarioRisk'],
			},
		},
		default: '',
		description: 'The ID of the risk to unlink',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAssessmentScenarioId = this.getNodeParameter('riskAssessmentScenarioId', itemIndex) as string;
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;

	const query = `
		mutation UnlinkRiskAssessmentScenarioRisk($input: UnlinkRiskAssessmentScenarioRiskInput!) {
			unlinkRiskAssessmentScenarioRisk(input: $input) {
				riskAssessmentScenario {
					id
					name
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { riskAssessmentScenarioId, riskId },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
