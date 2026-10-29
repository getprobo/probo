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

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { organizationLocator } from '../../locators';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest, readResourceId } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['task'],
		operation: ['updateLinearSettings'],
	}),
	{
		displayName: 'Default Team',
		name: 'defaultTeam',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateLinearSettings'],
			},
		},
		options: [
			{
				name: 'Keep',
				value: 'KEEP',
				description: 'Leave the default Linear team unchanged',
			},
			{
				name: 'Set',
				value: 'SET',
				description: 'Choose a Linear team as the organization default',
			},
			{
				name: 'Clear',
				value: 'CLEAR',
				description: 'Clear the default Linear team and turn issue import off',
			},
		],
		default: 'KEEP',
		description: 'Whether to change the default Linear team. Clearing it also turns issue import off.',
	},
	{
		displayName: 'Team ID',
		name: 'teamId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateLinearSettings'],
				defaultTeam: ['SET'],
			},
		},
		default: '',
		description: 'The Linear team ID to use as the organization default',
		required: true,
	},
	{
		displayName: 'Issue Import',
		name: 'issueImport',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateLinearSettings'],
			},
		},
		options: [
			{
				name: 'Keep',
				value: 'KEEP',
				description: 'Leave issue import unchanged',
			},
			{
				name: 'Active',
				value: 'ACTIVE',
				description: 'Create a Probo task for each new issue on the default Linear team and keep it in sync',
			},
			{
				name: 'Inactive',
				value: 'INACTIVE',
				description: 'Leave new issues on the default Linear team in Linear',
			},
		],
		default: 'KEEP',
		description: 'Whether new issues on the default Linear team are imported. Active requires a default team.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = readResourceId(this.getNodeParameter('organizationId', itemIndex));
	const defaultTeam = this.getNodeParameter('defaultTeam', itemIndex) as string;
	const issueImport = this.getNodeParameter('issueImport', itemIndex) as string;

	if (defaultTeam === 'KEEP' && issueImport === 'KEEP') {
		throw new NodeOperationError(this.getNode(), 'Choose a Linear team or an issue import status.');
	}

	const input: {
		organizationId: string;
		defaultTeamId?: string | null;
		issueImport?: string;
	} = { organizationId };

	if (defaultTeam === 'CLEAR') {
		input.defaultTeamId = null;
	} else if (defaultTeam === 'SET') {
		input.defaultTeamId = this.getNodeParameter('teamId', itemIndex) as string;
	}

	if (issueImport !== 'KEEP') {
		input.issueImport = issueImport;
	}

	const query = `
		mutation UpdateLinearSettings($input: UpdateLinearSettingsInput!) {
			updateLinearSettings(input: $input) {
				organization {
					id
					linearSettings {
						defaultTeam {
							id
							name
							key
						}
						issueImport
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
